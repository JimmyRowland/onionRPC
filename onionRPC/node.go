package onionRPC

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	fchecker "cs.ubc.ca/cpsc416/onionRPC/fcheck"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/relayNode"
	"github.com/google/uuid"
)

const (
	GUARD_NODE_TYPE = "Guard"
	EXIT_NODE_TYPE  = "Exit"
	RELAY_NODE_TYPE = "Relay"
)

type OnionNodeJoinRequest struct {
	Timestamp        time.Time
	FcheckAddr       string // Address the server will use to listen for fcheck connection
	ClientListenAddr string // Address the server will use to listen for client->server connections
	ServerListenAddr string // Address the server will use to listen for server->server connections
	NodeId           uuid.UUID
}

type OnionNodeJoinResponse struct {
	Timestamp time.Time
	Role      string
}

type NodeConfig struct {
	NodeId            string
	CoordAddr         string
	FcheckAddr        string
	ClientListenAddr  string
	ServerListenAddr  string
	TracingServerAddr string
	TracingIdentity   string
}

type Node struct {
	mu          sync.Mutex
	Id          uuid.UUID
	fc          *fchecker.Fcheck
	fcAddr      string
	NodeType    string
	NodeConfig  NodeConfig
	sessionKeys map[string]string
	RelayNode   relayNode.Node
	ExitNode    exitNode.Node
	GuardNode   guardNode.Node
}

func NewNode() *Node {
	id, err := uuid.NewRandom()
	checkErr(err, "Failed to generate UUID for onion node")
	return &Node{Id: id}
}

func (n *Node) Start(config NodeConfig) error {
	n.mu.Lock()
	n.NodeConfig = config

	// 1. Start fcheck
	fc := fchecker.NewFcheck()
	n.fc = &fc
	n.fcAddr = newLocalAddr()
	fcConfig := fchecker.StartStruct{
		AckLocalIPAckLocalPort: n.fcAddr,
	}
	n.fc.Start(fcConfig)

	n.mu.Unlock()

	// 2. Establish coordinator connection and get role
	n.connectToCoord()

	// 3. Adopt role and begin responding to requests
	switch n.NodeType {
	case GUARD_NODE_TYPE:
		go n.GuardNode.Start()
	case EXIT_NODE_TYPE:
		go n.ExitNode.Start()
	case RELAY_NODE_TYPE:
		go n.RelayNode.Start()
	default:
		return nil
	}

	return nil
}

func (n *Node) connectToCoord() {
	n.mu.Lock()
	conn, err := net.Dial("tcp", n.NodeConfig.CoordAddr)
	checkErr(err, "Failed to connect to coordinator")
	defer conn.Close()

	// Send join request
	msg := OnionNodeJoinRequest{
		time.Now(),
		n.NodeConfig.FcheckAddr,
		n.NodeConfig.ClientListenAddr,
		n.NodeConfig.ServerListenAddr,
		n.Id,
	}

	conn.Write(encode(msg))

	// Receive response
	recvbuf := make([]byte, 1024)
	_, err = conn.Read(recvbuf)
	checkErr(err, "Failed to read NodeJoinMessage from TCP connection\n")

	// Decode response
	tmpbuff := bytes.NewBuffer(recvbuf)
	res := new(OnionNodeJoinResponse)
	decoder := gob.NewDecoder(tmpbuff)
	decoder.Decode(res)

	// Adopt assigned role
	n.NodeType = res.Role
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
