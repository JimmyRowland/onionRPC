package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 4; i++ {
		listenAddr := fmt.Sprintf(":1000%d", i)
		node := onionRPC.Node{GuardNode: guardNode.Node{RoleConfig: role.RoleConfig{ListenAddr: listenAddr}}}
		go node.Start()
		time.Sleep(time.Second)
		client := onionRPC.Client{Guard: onionRPC.OnionNode{RpcAddress: listenAddr}}
		client.Start()
	}

	ch := make(chan bool)
	<-ch
}
