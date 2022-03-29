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

type ServerFail struct {
	ServerId uint8
}

type ServerFailHandledRecvd struct {
	FailedServerId   uint8
	AdjacentServerId uint8
}

type ServerJoiningRecvd struct {
	ServerId uint8
}

type ServerJoinedRecvd struct {
	ServerId uint8
}

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
}

func NewCoord() *Coord {
	return &Coord{}
}

func (c *Coord) Start(clientAPIListenAddr string, serverAPIListenAddr string, lostMsgsThresh uint8, ctrace *tracing.Tracer) error {
	c.mu.Lock()
	c.isActive = true
	c.mu.Unlock()

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

		recvbuf := make([]byte, 1024)
		_, err = conn.Read(recvbuf)
		checkErr(err, "Failed to read NodeJoinMessage from TCP connection\n")
		tmpbuff := bytes.NewBuffer(recvbuf)
		msg := new(onionRPC.OnionNodeJoinRequest)
		decoder := gob.NewDecoder(tmpbuff)
		decoder.Decode(msg)

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
		res := onionRPC.OnionNodeJoinResponse{time.Now(), nodeType}
		conn.Write(encode(res))
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
		c.mu.Lock()

		failedServerAddr := fail.UDPIpPort
		c.fc.StopMonitoring(failedServerAddr)

		var node NodeConnection
		for _, n := range c.guardNodes {
			if n.FcheckAddr == failedServerAddr {
				node = n
				c.nActiveGuards -= 1
			}
		}
		for _, n := range c.exitNodes {
			if n.FcheckAddr == failedServerAddr {
				node = n
				c.nActiveExits -= 1
			}
		}
		for _, n := range c.relayNodes {
			if n.FcheckAddr == failedServerAddr {
				node = n
				c.nActiveRelays -= 1
			}
		}

		node.IsActive = false

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

		if c.nActiveExits > 0 && c.nActiveGuards > 0 && c.nActiveRelays > 0 {
			// Randomly select guard, exit, and relay nodes
			guardIndex := rand.Intn(c.nActiveGuards)
			exitIndex := rand.Intn(c.nActiveExits)
			relayIndex := rand.Intn(c.nActiveRelays)
			var guard, exit, relay onionRPC.OnionNode

			for _, node := range c.guardNodes {
				if node.IsActive && guardIndex == 0 {
					guard = onionRPC.OnionNode{
						RpcAddress: node.ClientListenAddr,
					}
					break
				} else if node.IsActive {
					guardIndex -= 1
				}
			}
			for _, node := range c.exitNodes {
				if node.IsActive && guardIndex == 0 {
					exit = onionRPC.OnionNode{
						RpcAddress: node.ClientListenAddr,
					}
					break
				} else if node.IsActive {
					exitIndex -= 1
				}
			}
			for _, node := range c.relayNodes {
				if node.IsActive && guardIndex == 0 {
					relay = onionRPC.OnionNode{
						RpcAddress: node.ClientListenAddr,
					}
					break
				} else if node.IsActive {
					relayIndex -= 1
				}
			}
			conn.Write(encode(onionRPC.OnionCircuitResponse{
				Error:  nil,
				Guard:  guard,
				Exit:   exit,
				Relay:  relay,
				Relays: nil,
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
