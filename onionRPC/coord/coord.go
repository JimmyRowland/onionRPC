package coord

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	fchecker "cs.ubc.ca/cpsc416/onionRPC/fcheck"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"github.com/DistributedClocks/tracing"
	"github.com/google/uuid"
)

// Actions to be recorded by coord (as part of ctrace, ktrace, and strace):

type CoordStart struct {
}

type NodeFail struct {
}

type NodeFailHandledRecvd struct {
}

type OnionReqRecvd struct {
	ClientId     string
	OldExitAddr  string
	OldRelayAddr string
	OldGuardAddr string
}

type OnionRes struct {
	ClientId  string
	ExitAddr  string
	RelayAddr string
	GuardAddr string
}

type NodeJoiningRecvd struct {
	NodeId uuid.UUID
}

type NodeJoinedRecvd struct {
	NodeId uuid.UUID
	Role   string
}

//***

type CoordConfig struct {
	ClientAPIListenAddr string
	ServerAPIListenAddr string
	LostMsgsThresh      uint8
	TracingServerAddr   string
	Secret              []byte
	TracingIdentity     string
}

type Coord struct {
	// Coord state may go here
	mu            sync.Mutex
	fc            *fchecker.Fcheck
	stopChan      chan bool
	isActive      bool
	guardNodes    []NodeConnection
	exitNodes     []NodeConnection
	relayNodes    []NodeConnection
	nActiveGuards int
	nActiveExits  int
	nActiveRelays int
	tracer        *tracing.Tracer
}

func NewCoord() *Coord {
	return &Coord{}
}

func (c *Coord) Start(clientAPIListenAddr string, serverAPIListenAddr string, lostMsgsThresh uint8, ctrace *tracing.Tracer) error {
	c.mu.Lock()
	c.isActive = true
	c.mu.Unlock()

	c.tracer = ctrace

	trace := c.tracer.CreateTrace()
	trace.RecordAction(CoordStart{})

	// 1. Begin fcheck
	fcStruct := fchecker.StartStruct{
		HBeatLocalIPHBeatLocalPort: newLocalAddr(),
	}
	fc := fchecker.NewFcheck()
	c.fc = &fc
	notifyCh, err := c.fc.Start(fcStruct)
	checkErr(err, "Failed to start fcheck")

	// 2. Listen for node connections
	go c.handleNodeConnections(serverAPIListenAddr, lostMsgsThresh)

	// 3. Handle node failures
	go c.handleNodeFailures(notifyCh)

	// 4. Listen for client connections
	go c.handleClientConnections(clientAPIListenAddr)

	// 5. Wait for Stop command
	c.stopChan = make(chan bool)
	go c.awaitStop(c.stopChan)

	select {} // TODO: temporary - remove later
	return nil
}

func (c *Coord) Stop() {
	c.stopChan <- true
}

func (c *Coord) awaitStop(stopChan <-chan bool) {
	stop := <-stopChan
	if stop {
		// Shut down coord
		c.mu.Lock()
		c.isActive = false
		c.mu.Unlock()
	}
}

type NodeConnection struct {
	NodeId           uuid.UUID
	Type             string
	FcheckAddr       string
	ServerListenAddr string
	ClientListenAddr string
	IsActive         bool
}

func (c *Coord) handleNodeConnections(serverAPIListenAddr string, lostMsgsThresh uint8) {
	listener, err := net.Listen("tcp", serverAPIListenAddr)
	checkErr(err, "Coord failed to listen for TCP connections on ServerAPIListenAddr\n")

	for {
		conn, err := listener.Accept()
		checkErr(err, "Failed to accept connection from onion node on TCP socket\n")

		fmt.Println("Received connection from node.")
		recvbuf := make([]byte, 1024)
		_, err = conn.Read(recvbuf)
		checkErr(err, "Failed to read NodeJoinMessage from TCP connection\n")
		tmpbuff := bytes.NewBuffer(recvbuf)
		msg := new(onionRPC.OnionNodeJoinRequest)
		decoder := gob.NewDecoder(tmpbuff)
		decoder.Decode(msg)

		trace := c.tracer.ReceiveToken(msg.Token)
		trace.RecordAction(NodeJoiningRecvd{msg.NodeId})

		// Begin fcheck monitoring of node
		c.startMonitoringServer(msg.FcheckAddr, lostMsgsThresh)

		c.mu.Lock()

		nodeType := c.chooseNodeType()

		// Add node to roster
		nodeConn := NodeConnection{
			NodeId:           msg.NodeId,
			FcheckAddr:       msg.FcheckAddr,
			ServerListenAddr: msg.ServerListenAddr,
			ClientListenAddr: msg.ClientListenAddr,
			IsActive:         true,
		}
		didReplace := false
		var nodeArr *[]NodeConnection
		if nodeType == onionRPC.GUARD_NODE_TYPE {
			nodeArr = &c.guardNodes
			c.nActiveGuards += 1
		} else if nodeType == onionRPC.EXIT_NODE_TYPE {
			nodeArr = &c.exitNodes
			c.nActiveExits += 1
		} else {
			nodeArr = &c.relayNodes
			c.nActiveRelays += 1
		}
		// Look for node in roster, update it if present
		for _, node := range *nodeArr {
			if node.NodeId == nodeConn.NodeId {
				node.FcheckAddr = nodeConn.FcheckAddr
				node.ServerListenAddr = nodeConn.ServerListenAddr
				node.ClientListenAddr = nodeConn.ClientListenAddr
				node.IsActive = true

				didReplace = true
			}
		}
		// Otherwise, add the new node to the roster
		if !didReplace {
			*nodeArr = append(*nodeArr, nodeConn)
		}
		c.mu.Unlock()

		// Send acknowledgement to node
		trace.RecordAction(NodeJoinedRecvd{NodeId: msg.NodeId, Role: nodeType})
		res := onionRPC.OnionNodeJoinResponse{time.Now(), nodeType, trace.GenerateToken()}
		conn.Write(encode(res))
		if !didReplace {
			fmt.Println("Established connection with a new onion node")
		} else {
			fmt.Println("Re-established connection with a failed onion node")
		}
		conn.Close()
	}

}

// This function assumes that the caller is holding the lock for the mutex `c.mu`
func (c *Coord) chooseNodeType() string {
	// 1. Prioritoze getting 1 of each type
	if c.nActiveGuards == 0 {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.nActiveExits == 0 {
		return onionRPC.EXIT_NODE_TYPE
	} else if c.nActiveRelays == 0 {
		return onionRPC.RELAY_NODE_TYPE
	}

	// 2. Prioritize getting at least 2 guards and exits
	if c.nActiveGuards < 2 {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.nActiveExits < 2 {
		return onionRPC.EXIT_NODE_TYPE
	}

	DESIRED_RATIO := 3 // Ratio of relay nodes to guard/exit nodes

	// 3. Finally, attempt to achieve the desired ratio of relays to guards/exits
	if c.nActiveRelays/c.nActiveGuards > DESIRED_RATIO {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.nActiveRelays/c.nActiveExits > DESIRED_RATIO {
		return onionRPC.EXIT_NODE_TYPE
	}

	return onionRPC.RELAY_NODE_TYPE
}

func (c *Coord) handleNodeFailures(notifyCh <-chan fchecker.FailureDetected) {
	c.mu.Lock()
	isCoordActive := c.isActive
	c.mu.Unlock()
	for isCoordActive {
		fail := <-notifyCh
		fmt.Println("Onion coordinator detected node failure")
		c.mu.Lock()

		trace := c.tracer.CreateTrace()

		failedServerAddr := fail.UDPIpPort
		c.fc.StopMonitoring(failedServerAddr)

		trace.RecordAction(NodeFail{})

		var node *NodeConnection
		for i := range c.guardNodes {
			if c.guardNodes[i].FcheckAddr == failedServerAddr {
				node = &c.guardNodes[i]
				c.nActiveGuards -= 1
			}
		}
		for i := range c.exitNodes {
			if c.exitNodes[i].FcheckAddr == failedServerAddr {
				node = &c.exitNodes[i]
				c.nActiveExits -= 1
			}
		}
		for i := range c.relayNodes {
			if c.relayNodes[i].FcheckAddr == failedServerAddr {
				node = &c.relayNodes[i]
				c.nActiveRelays -= 1
			}
		}

		node.IsActive = false

		fmt.Println("Node failure was processed by the coordinator")
		trace.RecordAction(NodeFailHandledRecvd{})

		c.mu.Unlock()
	}
}

func (c *Coord) handleClientConnections(clientAPIListenAddr string) {
	listener, err := net.Listen("tcp", clientAPIListenAddr)
	checkErr(err, "Coord failed to listen for TCP connections on ClientAPIListenAddr\n")

	for {
		conn, err := listener.Accept()
		checkErr(err, "Failed to accept connection from onion client on TCP socket\n")

		// Decode request
		recvbuf := make([]byte, 1024)
		_, err = conn.Read(recvbuf)
		checkErr(err, "Failed to read OnionCircuitRequest from TCP connection\n")
		tmpbuff := bytes.NewBuffer(recvbuf)
		msg := new(onionRPC.OnionCircuitRequest)
		decoder := gob.NewDecoder(tmpbuff)
		decoder.Decode(msg)

		c.mu.Lock()

		trace := c.tracer.ReceiveToken(msg.Token)
		trace.RecordAction(OnionReqRecvd{
			ClientId:     msg.ClientID,
			OldExitAddr:  msg.OldExitAddr,
			OldGuardAddr: msg.OldGuardAddr,
			OldRelayAddr: msg.OldRelayAddr})

		if c.nActiveExits > 0 && c.nActiveGuards > 0 && c.nActiveRelays > 0 {
			// Randomly select guard, exit, and relay nodes
			guard, exit, relay := c.getNewCircuit(msg.OldGuardAddr, msg.OldRelayAddr, msg.OldExitAddr)
			trace.RecordAction(OnionRes{
				ClientId:  msg.ClientID,
				ExitAddr:  exit.RpcAddress,
				RelayAddr: relay.RpcAddress,
				GuardAddr: guard.RpcAddress,
			})
			conn.Write(encode(onionRPC.OnionCircuitResponse{
				Error:  nil,
				Guard:  guard,
				Exit:   exit,
				Relay:  relay,
				Relays: nil,
				Token:  trace.GenerateToken(),
			}))
		} else {
			conn.Write(encode(onionRPC.OnionCircuitResponse{
				Error: fmt.Errorf("Insufficient nodes in network to generate a valid onion circuit"),
			}))
		}
		conn.Close()

		c.mu.Unlock()
	}
}

// Randomly generates a new circuit
// Assumes caller is hold c.mutex
// Assumes coord has at least 1 active guard, relay, and exit node
func (c *Coord) getNewCircuit(oldGuardAddr, oldRelayAddr, oldExitAddr string) (guard, exit, relay onionRPC.OnionNode) {

	getActiveValue := func(arr []NodeConnection, nActive int, oldAddr string) *NodeConnection {
		var retVal *NodeConnection
		siz := len(arr)
		idx := rand.Intn(siz)
		for {
			idx %= siz
			if nActive > 1 {
				if arr[idx].IsActive && arr[idx].ClientListenAddr != oldAddr {
					retVal = &arr[idx]
					return retVal
				}
			} else {
				if arr[idx].IsActive {
					retVal = &arr[idx]
					return retVal
				}
			}

			idx++
		}
	}
	guardNode := getActiveValue(c.guardNodes, c.nActiveGuards, oldGuardAddr)
	exitNode := getActiveValue(c.exitNodes, c.nActiveExits, oldExitAddr)
	relayNode := getActiveValue(c.relayNodes, c.nActiveRelays, oldRelayAddr)

	_func := func(c *NodeConnection) onionRPC.OnionNode {
		return onionRPC.OnionNode{RpcAddress: c.ClientListenAddr}
	}
	return _func(guardNode), _func(exitNode), _func(relayNode)
}

func (c *Coord) startMonitoringServer(raddr string, lostMsgsThresh uint8) {
	thresh := int(lostMsgsThresh)
	err := c.fc.BeginMonitoring(newLocalAddr(), raddr, thresh)
	checkErr(err, "Failed to begin fcheck heartbeat for new server\n")
}

func newLocalAddr() string {
	port, err := getFreePort()
	checkErr(err, "Failed to find a free local port")
	return "0.0.0.0:" + port
}

func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer l.Close()
	return fmt.Sprint(l.Addr().(*net.TCPAddr).Port), nil
}

func checkErr(err error, errfmsg string, fargs ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, errfmsg, fargs...)
		os.Exit(1)
	}
}

// Encodes arbitrary data using gob
func encode(message interface{}) []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(message)
	return buf.Bytes()
}
