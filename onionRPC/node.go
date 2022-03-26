package onionRPC

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
)

const (
	GUARD_NODE_TYPE = "guard"
	EXIT_NODE_TYPE  = "exit"
	RELAY_NODE_TYPE = "relay"
)

type NodeConfig struct {
	NodeId                 int
	CoordAddr              string
	AckLocalIPAckLocalPort string
	TracingServerAddr      string
	TracingIdentity        string
}

type Node struct {
	NodeType    string
	NodeConfig  NodeConfig
	sessionKeys map[string]string
	RelayNode   guardNode.Node
	ExitNode    guardNode.Node
	GuardNode   guardNode.Node
}

func (node *Node) Start() error {
	//Get role from coordinator
	node.NodeType = GUARD_NODE_TYPE
	if node.NodeType == GUARD_NODE_TYPE {
		go node.GuardNode.Start()
	}
	return nil
}
