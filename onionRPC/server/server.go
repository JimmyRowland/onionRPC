package server

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	"cs.ubc.ca/cpsc416/onionRPC/util"
	"github.com/DistributedClocks/tracing"
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

type GetRandNum struct{ Number int }

type GetLongString struct {
	LongString string
	Number     int
}

type Request struct {
	Token tracing.TracingToken
	Args  Operands
}

type Response struct {
	Token tracing.TracingToken
	Res   Result
}

type Operands struct {
	A int
	B int
}

type Result struct {
	Result int
	Token  tracing.TracingToken
}

func (server *Server) Add(operands Request, result *Response) error {
	trace := server.tracer.ReceiveToken(operands.Token)
	trace.RecordAction(operands)
	result.Res.Result = operands.Args.A + operands.Args.B
	trace.RecordAction(*result)
	result.Token = trace.GenerateToken()
	return nil
}

func (server *Server) Subtract(operands Request, result *Response) error {
	trace := server.tracer.ReceiveToken(operands.Token)
	trace.RecordAction(operands)
	result.Res.Result = operands.Args.A - operands.Args.B
	trace.RecordAction(*result)
	result.Token = trace.GenerateToken()
	return nil
}

func (server *Server) Multiply(operands Request, result *Response) error {
	trace := server.tracer.ReceiveToken(operands.Token)
	trace.RecordAction(operands)
	result.Res.Result = operands.Args.A * operands.Args.B
	trace.RecordAction(*result)
	result.Token = trace.GenerateToken()
	return nil
}

func (server *Server) Divide(operands Request, result *Response) error {
	trace := server.tracer.ReceiveToken(operands.Token)
	trace.RecordAction(operands)
	if operands.Args.B == 0 {
		return errors.New("division by 0")
	}
	result.Res.Result = operands.Args.A / operands.Args.B
	trace.RecordAction(*result)
	result.Token = trace.GenerateToken()
	return nil
}

func (server *Server) GetRandomNumber(_ RandomNumber, randomNumber *RandomNumber) error {
	randomNumber.Number = rand.Intn(10000)
	server.trace.RecordAction(GetRandNum{Number: randomNumber.Number})
	fmt.Println("Generating random number: ", randomNumber.Number)
	return nil
}

func (server *Server) GetLongString(_ LongString, longString *LongString) error {
	longString.String = "qwertyuioplkjhgdsazxcvbnmqwertyuioplkjhgfdsazxcvbnmqwertyuioplkjhgdsazxcvbnmqwertyuioplkjhgfdsazxcvbnm"
	longString.Number = rand.Intn(10000)
	server.trace.RecordAction(GetLongString{LongString: longString.String, Number: longString.Number})
	fmt.Println(longString.String, longString.Number)
	return nil
}

func (server *Server) Start(config Config, ctrace *tracing.Tracer) error {
	server.tracer = ctrace
	server.trace = ctrace.CreateTrace()
	server.trace.RecordAction(ServerStart{config})
	l, err := net.Listen("tcp", config.ServerAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	err = rpc.Register(server)
	util.CheckErr(err, config.ServerAddr)
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
