package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"github.com/google/uuid"
	"time"
)

func main() {
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
	time.Sleep(time.Second)

	go node2.Start(nodeConfig2)
	time.Sleep(time.Second)

	go node3.Start(nodeConfig3)
	time.Sleep(time.Second)

	select {}
}
