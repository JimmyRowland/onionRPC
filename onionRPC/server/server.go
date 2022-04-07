package server

import (
	"context"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"google.golang.org/grpc"
	"math/rand"
	"net"
)

type Server struct {
	Config Config
	UnimplementedServerServiceServer
}

type Config struct {
	ServerID          int
	ServerAddr        string
	TracingServerAddr string
	Secret            string
	TracingIdentity   string
}

func (server *Server) GetRandomNumber(_ context.Context, _ *RandomNumber) (*RandomNumber, error) {
	num := RandomNumber{Number: int32(rand.Uint32())}
	fmt.Println("Generating random number: ", num.Number)
	return &num, nil
}

func (server *Server) Start(config Config) error {
	lis, err := net.Listen("tcp", config.ServerAddr)
	util.CheckErr(err, config.ServerAddr)
	defer lis.Close()

	grpcServer := grpc.NewServer()
	RegisterServerServiceServer(grpcServer, server)
	fmt.Println("Server started ", config.ServerAddr)
	err = grpcServer.Serve(lis)
	util.CheckErr(err, config.ServerAddr)

	select {}
}
