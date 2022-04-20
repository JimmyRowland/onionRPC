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
	"math/rand"
	"sync"
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

func startCoord() *coord.Coord {
	var config coord.CoordConfig
	util.ReadJSONConfig("../config/coord_config.json", &config)
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	coord := coord.NewCoord()
	go coord.Start(config.ClientAPIListenAddr, config.ServerAPIListenAddr, config.LostMsgsThresh, ctracer)
	return coord
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
		CoordAddr:         ":1235",
		FcheckAddr:        addIdToString(id, ":500"),
		ClientListenAddr:  addIdToString(id, ":501"),
		TracingServerAddr: ":56666",
		TracingIdentity:   addIdToString(id, "node"),
	}
	node := onionRPC.NewNode()
	node.Start(nodeConfig)
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

func startClientWithId(clientId string) *onionRPC.Client {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("../config/client_config.json", &config)
	config.ClientID = clientId
	config.TracingIdentity = clientId
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

func TestOnionFailureCoordHandling(t *testing.T) {
	startTracer()
	startCoord()
	guard1 := startNode(0)
	exit1 := startNode(1)
	relay1 := startNode(2)
	guard1.Close()
	exit1.Close()
	relay1.Close()
	startNode(3)
	startNode(4)
	startNode(5)
	time.Sleep(time.Second * 2)
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
	startNode(4)
	startNode(5)
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
	startNode(3)
	startNode(4)
	startNode(5)
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

func TestOnionChainWithAsyncFailed(t *testing.T) {
	startTracer()
	coord := startCoord()
	coord.DESIRED_RATIO = 1
	mockServer := startMockServer()
	nodeMap := make(map[string][]*onionRPC.Node)
	nodeMap[onionRPC.GUARD_NODE_TYPE] = append(nodeMap[onionRPC.GUARD_NODE_TYPE], startNode(0))
	nodeMap[onionRPC.EXIT_NODE_TYPE] = append(nodeMap[onionRPC.EXIT_NODE_TYPE], startNode(1))
	nodeMap[onionRPC.RELAY_NODE_TYPE] = append(nodeMap[onionRPC.RELAY_NODE_TYPE], startNode(2))

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	runClientTests := func(clientId string) {
		client := startClientWithId(clientId)
		operands := Operands{A: 1, B: 1}
		result := Result{}
		for i := 0; i < 10; i++ {
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
		waitGroup.Done()
	}
	go runClientTests("client1")
	go runClientTests("client2")
	go func() {
		var idMux sync.Mutex
		var id uint8
		id = 3
		closeNodes := func(nodeType string) {
			for i := 0; i < 10; i++ {
				idMux.Lock()
				nodes := nodeMap[nodeType]
				//fmt.Println(nodeType, len(nodes))
				//fmt.Println("guard", len(coord.GuardNodes), " relay:", len(coord.RelayNodes), " exit:", len(coord.ExitNodes))
				//fmt.Println("guard", coord.NActiveGuards, " relay:", coord.NActiveRelays, " exit:", coord.NActiveExits)
				if len(nodes) > 0 {
					randNode := rand.Intn(len(nodes)) - 1
					for i := 0; i < len(nodes); i++ {
						if randNode == i {
							continue
						}
						nodes[i].Close()
					}
				}

				nodeMap[nodeType] = []*onionRPC.Node{}
				idMux.Unlock()
				for i := 0; i < 3; i++ {
					idMux.Lock()
					node := startNode(id)
					switch node.NodeType {
					case onionRPC.GUARD_NODE_TYPE:
						nodeMap[onionRPC.GUARD_NODE_TYPE] = append(nodeMap[onionRPC.GUARD_NODE_TYPE], node)
					case onionRPC.RELAY_NODE_TYPE:
						nodeMap[onionRPC.RELAY_NODE_TYPE] = append(nodeMap[onionRPC.RELAY_NODE_TYPE], node)
					case onionRPC.EXIT_NODE_TYPE:
						nodeMap[onionRPC.EXIT_NODE_TYPE] = append(nodeMap[onionRPC.EXIT_NODE_TYPE], node)
					}
					id = (id + 1) % 99
					idMux.Unlock()
				}

				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)+5000))
			}
		}
		go closeNodes(onionRPC.GUARD_NODE_TYPE)
		go closeNodes(onionRPC.RELAY_NODE_TYPE)
		go closeNodes(onionRPC.EXIT_NODE_TYPE)

	}()
	waitGroup.Wait()
}
