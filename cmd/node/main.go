package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/relayNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"fmt"
	"time"
)

type RandomNumber struct {
	number int
}

func main() {
	for i := 0; i < 4; i++ {
		guardListenAddr := fmt.Sprintf(":1000%d", i)
		guard := onionRPC.Node{NodeType: onionRPC.GUARD_NODE_TYPE, GuardNode: guardNode.Node{RoleConfig: role.RoleConfig{ListenAddr: guardListenAddr}}}
		go guard.Start()
		relayListenAddr := fmt.Sprintf(":1001%d", i)
		relay := onionRPC.Node{NodeType: onionRPC.RELAY_NODE_TYPE, RelayNode: relayNode.Node{RoleConfig: role.RoleConfig{ListenAddr: relayListenAddr}}}
		go relay.Start()
		exitListenAddr := fmt.Sprintf(":1002%d", i)
		exit := onionRPC.Node{NodeType: onionRPC.EXIT_NODE_TYPE, ExitNode: exitNode.Node{RoleConfig: role.RoleConfig{ListenAddr: exitListenAddr}}}
		go exit.Start()
		time.Sleep(time.Second)
		client := onionRPC.Client{Guard: onionRPC.OnionNode{RpcAddress: guardListenAddr}, Relay: onionRPC.OnionNode{RpcAddress: relayListenAddr}, Exit: onionRPC.OnionNode{RpcAddress: exitListenAddr}, ClientConfig: onionRPC.ClientConfig{ClientID: "string(i)"}}
		client.Start()
		client.RpcCall("server", "Server.GetRandomNumber", RandomNumber{number: 2}, &RandomNumber{number: 111})
	}

	ch := make(chan bool)
	<-ch
}
