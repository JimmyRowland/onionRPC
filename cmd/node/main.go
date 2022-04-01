package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"github.com/google/uuid"
)

type RandomNumber struct {
	number int
}

func main() {
	for i := 0; i < 4; i++ {

		nodeConfig1 := onionRPC.NodeConfig{
			NodeId:            "node1",
			CoordAddr:         "127.0.0.1:1235",
			FcheckAddr:        "127.0.0.1:4040",
			ClientListenAddr:  "127.0.0.1:4050",
			ServerListenAddr:  "127.0.0.1:4060",
			TracingServerAddr: "127.0.0.1:6666",
			TracingIdentity:   "node1",
		}
		nodeConfig2 := onionRPC.NodeConfig{
			NodeId:            "node2",
			CoordAddr:         "127.0.0.1:1235",
			FcheckAddr:        "127.0.0.1:4041",
			ClientListenAddr:  "127.0.0.1:4051",
			ServerListenAddr:  "127.0.0.1:4061",
			TracingServerAddr: "127.0.0.1:6666",
			TracingIdentity:   "node2",
		}
		nodeConfig3 := onionRPC.NodeConfig{
			NodeId:            "node3",
			CoordAddr:         "127.0.0.1:1235",
			FcheckAddr:        "127.0.0.1:4042",
			ClientListenAddr:  "127.0.0.1:4052",
			ServerListenAddr:  "127.0.0.1:4062",
			TracingServerAddr: "127.0.0.1:6666",
			TracingIdentity:   "node3",
		}
		node1 := onionRPC.NewNode()
		node2 := onionRPC.NewNode()
		node3 := onionRPC.NewNode()
		// TODO: overwriting ID fields for testing purposes
		id1, err := uuid.FromBytes([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
		id2, err := uuid.FromBytes([]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})
		id3, err := uuid.FromBytes([]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3})
		node1.Id = id1
		node2.Id = id2
		node3.Id = id3
		if err != nil {
			panic("AHH")
		}
		go node1.Start(nodeConfig1)
		go node2.Start(nodeConfig2)
		go node3.Start(nodeConfig3)
		select {}
		// guardListenAddr := fmt.Sprintf(":1000%d", i)
		// guard := onionRPC.Node{NodeType: onionRPC.GUARD_NODE_TYPE, GuardNode: guardNode.Node{RoleConfig: role.RoleConfig{ListenAddr: guardListenAddr}}}
		// go guard.Start()
		// relayListenAddr := fmt.Sprintf(":1001%d", i)
		// relay := onionRPC.Node{NodeType: onionRPC.RELAY_NODE_TYPE, RelayNode: relayNode.Node{RoleConfig: role.RoleConfig{ListenAddr: relayListenAddr}}}
		// go relay.Start()
		// exitListenAddr := fmt.Sprintf(":1002%d", i)
		// exit := onionRPC.Node{NodeType: onionRPC.EXIT_NODE_TYPE, ExitNode: exitNode.Node{RoleConfig: role.RoleConfig{ListenAddr: exitListenAddr}}}
		// go exit.Start()
		// time.Sleep(time.Second)
		// client := onionRPC.Client{Guard: onionRPC.OnionNode{RpcAddress: guardListenAddr}, Relay: onionRPC.OnionNode{RpcAddress: relayListenAddr}, Exit: onionRPC.OnionNode{RpcAddress: exitListenAddr}, ClientConfig: onionRPC.ClientConfig{ClientID: "string(i)"}}
		// client.Start()
		// client.RpcCall("server", "Server.GetRandomNumber", RandomNumber{number: 2}, &RandomNumber{number: 111})

	}

	ch := make(chan bool)
	<-ch
}
