package coord

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
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
	NodeId uuid.UUID
	Type   string
}

type NodeFailHandled struct {
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
	GuardNodes    []NodeConnection
	ExitNodes     []NodeConnection
	RelayNodes    []NodeConnection
	NActiveGuards int
	NActiveExits  int
	NActiveRelays int
	DESIRED_RATIO int
	tracer        *tracing.Tracer
	trace         *tracing.Trace
}

func NewCoord() *Coord {
	return &Coord{}
}

func (c *Coord) Start(clientAPIListenAddr string, serverAPIListenAddr string, lostMsgsThresh uint8, ctrace *tracing.Tracer) error {
	c.mu.Lock()
	c.isActive = true
	c.mu.Unlock()

	c.tracer = ctrace

	c.trace = c.tracer.CreateTrace()
	c.trace.RecordAction(CoordStart{})

	// 1. Listen for node connections
	go c.handleNodeConnections(serverAPIListenAddr, lostMsgsThresh)

	// 2. Listen for client connections
	go c.handleClientConnections(clientAPIListenAddr)

	// 3. Wait for Stop command
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

		ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
		msg.FcheckAddr = ip + ":" + strings.Split(msg.FcheckAddr, ":")[1]
		msg.ClientListenAddr = ip + ":" + strings.Split(msg.ClientListenAddr, ":")[1]
		// Begin fcheck monitoring of node
		go c.startMonitoringServer(msg.FcheckAddr, lostMsgsThresh)

		c.mu.Lock()

		nodeType := c.chooseNodeType()

		// Add node to roster
		nodeConn := NodeConnection{
			NodeId:           msg.NodeId,
			FcheckAddr:       msg.FcheckAddr,
			ServerListenAddr: msg.ServerListenAddr,
			ClientListenAddr: msg.ClientListenAddr,
			IsActive:         true,
			Type:             nodeType,
		}
		didReplace := false
		var nodeArr *[]NodeConnection
		if nodeType == onionRPC.GUARD_NODE_TYPE {
			nodeArr = &c.GuardNodes
			c.NActiveGuards += 1
		} else if nodeType == onionRPC.EXIT_NODE_TYPE {
			nodeArr = &c.ExitNodes
			c.NActiveExits += 1
		} else {
			nodeArr = &c.RelayNodes
			c.NActiveRelays += 1
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
	if c.NActiveGuards == 0 {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.NActiveExits == 0 {
		return onionRPC.EXIT_NODE_TYPE
	} else if c.NActiveRelays == 0 {
		return onionRPC.RELAY_NODE_TYPE
	}

	// 2. Prioritize getting at least 2 guards and exits
	if c.NActiveGuards < 2 {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.NActiveExits < 2 {
		return onionRPC.EXIT_NODE_TYPE
	}

	if c.DESIRED_RATIO == 0 {
		c.DESIRED_RATIO = 3
	} // Ratio of relay nodes to guard/exit nodes

	// 3. Finally, attempt to achieve the desired ratio of relays to guards/exits
	if c.NActiveRelays/c.NActiveGuards > c.DESIRED_RATIO {
		return onionRPC.GUARD_NODE_TYPE
	} else if c.NActiveRelays/c.NActiveExits > c.DESIRED_RATIO {
		return onionRPC.EXIT_NODE_TYPE
	}

	return onionRPC.RELAY_NODE_TYPE
}

func (c *Coord) handleNodeFailures(notifyCh <-chan fchecker.FailureDetected, fcheck *fchecker.Fcheck) {

	fail := <-notifyCh
	fcheck.Stop()
	fmt.Println("Onion coordinator detected node failure")
	c.mu.Lock()

	failedServerAddr := fail.UDPIpPort

	var node *NodeConnection
	for i := range c.GuardNodes {
		if c.GuardNodes[i].FcheckAddr == failedServerAddr {
			node = &c.GuardNodes[i]
			c.NActiveGuards -= 1
		}
	}
	for i := range c.ExitNodes {
		if c.ExitNodes[i].FcheckAddr == failedServerAddr {
			node = &c.ExitNodes[i]
			c.NActiveExits -= 1
		}
	}
	for i := range c.RelayNodes {
		if c.RelayNodes[i].FcheckAddr == failedServerAddr {
			node = &c.RelayNodes[i]
			c.NActiveRelays -= 1
		}
	}

	node.IsActive = false

	c.trace.RecordAction(NodeFail{NodeId: node.NodeId, Type: node.Type})

	c.trace.RecordAction(NodeFailHandled{})

	c.mu.Unlock()

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

		if c.NActiveExits > 0 && c.NActiveGuards > 0 && c.NActiveRelays > 0 {
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
	guardNode := getActiveValue(c.GuardNodes, c.NActiveGuards, oldGuardAddr)
	exitNode := getActiveValue(c.ExitNodes, c.NActiveExits, oldExitAddr)
	relayNode := getActiveValue(c.RelayNodes, c.NActiveRelays, oldRelayAddr)

	_func := func(c *NodeConnection) onionRPC.OnionNode {
		return onionRPC.OnionNode{RpcAddress: c.ClientListenAddr}
	}
	return _func(guardNode), _func(exitNode), _func(relayNode)
}

func (c *Coord) startMonitoringServer(raddr string, lostMsgsThresh uint8) {
	freeUDPAddrPort := newLocalAddr()
	var fcheck fchecker.Fcheck
	fcheckerError, err := fcheck.Start(fchecker.StartStruct{"", rand.Uint64(), freeUDPAddrPort, raddr, lostMsgsThresh})
	checkErr(err, "Failed to begin fcheck heartbeat for new server\n")

	go c.handleNodeFailures(fcheckerError, &fcheck)
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
