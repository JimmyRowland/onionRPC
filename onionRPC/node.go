package onionRPC

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/relayNode"
)

const (
	GUARD_NODE_TYPE = "Guard"
	EXIT_NODE_TYPE  = "Exit"
	RELAY_NODE_TYPE = "Relay"
)

type NodeConfig struct {
	NodeId                 string
	CoordAddr              string
	AckLocalIPAckLocalPort string
	TracingServerAddr      string
	TracingIdentity        string
}

type Node struct {
	NodeType    string
	NodeConfig  NodeConfig
	sessionKeys map[string]string
	RelayNode   relayNode.Node
	ExitNode    exitNode.Node
	GuardNode   guardNode.Node
}

func (node *Node) Start() error {
	//Get role from coordinator
	switch node.NodeType {
	case GUARD_NODE_TYPE:
		go node.GuardNode.Start()
	case EXIT_NODE_TYPE:
		go node.ExitNode.Start()
	case RELAY_NODE_TYPE:
		go node.RelayNode.Start()
	default:
		return nil
	}
	return nil
}
