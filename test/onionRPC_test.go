package test

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/coord"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/server"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"github.com/DistributedClocks/tracing"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

type Operands struct {
	A int
	B int
}

type Result struct {
	Result int
}

func startTracer() {
	tracingServer := tracing.NewTracingServerFromFile("../config/tracing_server_config.json")
	err := tracingServer.Open()
	if err != nil {
		log.Fatal(err)
	}
	go tracingServer.Accept()
	time.Sleep(time.Millisecond * 10)
}

func startCoord() {
	var config coord.CoordConfig
	util.ReadJSONConfig("../config/coord_config.json", &config)
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	coord := coord.NewCoord()
	go coord.Start(config.ClientAPIListenAddr, config.ServerAPIListenAddr, config.LostMsgsThresh, ctracer)
}

func addIdToString(id uint8, string string) string {
	if id < 10 {
		return fmt.Sprintf("%s0%d", string, id)
	}
	return fmt.Sprintf("%s%d", string, id)
}

func startNode(id uint8) *onionRPC.Node {
	nodeConfig := onionRPC.NodeConfig{
		NodeId:            addIdToString(id, "node"),
		CoordAddr:         "127.0.0.1:1235",
		FcheckAddr:        addIdToString(id, "127.0.0.1:500"),
		ClientListenAddr:  addIdToString(id, "127.0.0.1:501"),
		TracingServerAddr: "127.0.0.1:6666",
		TracingIdentity:   addIdToString(id, "node"),
	}
	node := onionRPC.NewNode()
	go node.Start(nodeConfig)
	time.Sleep(time.Millisecond * 10)
	return node
}

func startMockServer() *server.Server {
	var config server.Config
	util.ReadJSONConfig("../config/server_config.json", &config)
	_server := server.Server{Config: config}
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	go _server.Start(_server.Config, ctracer)
	return &_server
}

func startClient() *onionRPC.Client {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("../config/client_config.json", &config)
	client := onionRPC.Client{ClientConfig: config}
	client.Start(client.ClientConfig)
	return &client
}

func TestOnionChainWithoutFailure(t *testing.T) {
	startTracer()
	startCoord()
	mockServer := startMockServer()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	fmt.Println(guard1, exit1, relay1)
	client := startClient()
	operands := Operands{A: 1, B: 1}
	result := Result{}
	err := client.RpcCall(mockServer.Config.ServerAddr, "Server.Add", operands, &result)
	assert.Nil(t, err)
	assert.Equal(t, 2, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 8, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.Result)
}
func TestOnionChainWithFailedGuard(t *testing.T) {
	startTracer()
	startCoord()
	mockServer := startMockServer()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	fmt.Println(guard1, exit1, relay1)
	client := startClient()
	operands := Operands{A: 1, B: 1}
	result := Result{}
	err := client.RpcCall(mockServer.Config.ServerAddr, "Server.Add", operands, &result)
	assert.Nil(t, err)
	assert.Equal(t, 2, result.Result)
	guard1.Close()
	startNode(3)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 8, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.Result)
}

func TestOnionChainWithFailedRelay(t *testing.T) {
	startTracer()
	startCoord()
	mockServer := startMockServer()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	fmt.Println(guard1, exit1, relay1)
	client := startClient()
	operands := Operands{A: 1, B: 1}
	result := Result{}
	err := client.RpcCall(mockServer.Config.ServerAddr, "Server.Add", operands, &result)
	assert.Nil(t, err)
	assert.Equal(t, 2, result.Result)
	relay1.Close()
	time.Sleep(time.Second * 10)
	startNode(3)
	startNode(4)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 8, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.Result)
}
func TestOnionChainWithFailedExit(t *testing.T) {
	startTracer()
	startCoord()
	mockServer := startMockServer()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	fmt.Println(guard1, exit1, relay1)
	client := startClient()
	operands := Operands{A: 1, B: 1}
	result := Result{}
	err := client.RpcCall(mockServer.Config.ServerAddr, "Server.Add", operands, &result)
	assert.Nil(t, err)
	assert.Equal(t, 2, result.Result)
	startNode(3)
	startNode(4)
	startNode(5)
	exit1.Close()
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 8, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.Result)
}

func TestOnionChainWithFailedNodes(t *testing.T) {
	startTracer()
	startCoord()
	mockServer := startMockServer()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	fmt.Println(guard1, exit1, relay1)
	client := startClient()
	operands := Operands{A: 1, B: 1}
	result := Result{}
	err := client.RpcCall(mockServer.Config.ServerAddr, "Server.Add", operands, &result)
	assert.Nil(t, err)
	assert.Equal(t, 2, result.Result)
	startNode(4)
	startNode(5)
	startNode(6)
	guard1.Close()
	exit1.Close()
	relay1.Close()
	startNode(3)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 8, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result.Result)
	err = client.RpcCall(mockServer.Config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
	assert.Nil(t, err)
	assert.Equal(t, 1, result.Result)
}
