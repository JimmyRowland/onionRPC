package server

import (
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"github.com/DistributedClocks/tracing"
	"math/rand"
	"net"
	"net/http"
)

type Server struct {
	Config Config
	tracer *tracing.Tracer
	trace  *tracing.Trace
}

type Config struct {
	ServerID          int
	ServerAddr        string
	TracingServerAddr string
	Secret            []byte
	TracingIdentity   string
}

type RandomNumber struct {
	Number int
}

type LongString struct {
	String string
	Number int
}

type ServerStart struct{ Config Config }

func (server *Server) GetRandomNumber(_ RandomNumber, randomNumber *RandomNumber) error {
	randomNumber.Number = rand.Intn(10000)
	fmt.Println("Generating random number: ", randomNumber.Number)
	return nil
}

func (server *Server) GetLongString(_ LongString, longString *LongString) error {
	longString.String = "qwertyuioplkjhgdsazxcvbnmqwertyuioplkjhgfdsazxcvbnmqwertyuioplkjhgdsazxcvbnmqwertyuioplkjhgfdsazxcvbnm"
	longString.Number = rand.Intn(10000)
	fmt.Println(longString.String, longString.Number)
	return nil
}

func (server *Server) Start(config Config, ctrace *tracing.Tracer) error {
	server.tracer = ctrace
	server.trace = ctrace.CreateTrace()
	server.trace.RecordAction(ServerStart{config})
	rpcServer := NewRPCServer()
	err := rpcServer.Register(server)
	util.CheckErr(err, config.ServerAddr)
	lis, err := net.Listen("tcp", config.ServerAddr)
	util.CheckErr(err, config.ServerAddr)
	defer lis.Close()
	NewHTTPServer(rpcServer).Serve(lis)

	http.Serve(lis, nil)
	fmt.Println("RPCServer started ", config.ServerAddr)
	select {}
}
