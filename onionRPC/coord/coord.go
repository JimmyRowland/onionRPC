package coord

import (
	"context"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"google.golang.org/grpc"
	"net"
)

// Actions to be recorded by coord (as part of ctrace, ktrace, and strace):

type CoordStart struct {
}

type ServerFail struct {
	ServerId uint8
}

type ServerFailHandledRecvd struct {
	FailedServerId   uint8
	AdjacentServerId uint8
}

type NewChain struct {
	Chain []uint8
}

type AllServersJoined struct {
}

type HeadReqRecvd struct {
	ClientId string
}

type HeadRes struct {
	ClientId string
	ServerId uint8
}

type TailReqRecvd struct {
	ClientId string
}

type TailRes struct {
	ClientId string
	ServerId uint8
}

type ServerJoiningRecvd struct {
	ServerId uint8
}

type ServerJoinedRecvd struct {
	ServerId uint8
}

type CoordConfig struct {
	ListenAddr        string
	TracingServerAddr string
	TracingIdentity   string
}

type Coord struct {
	CoordConfig CoordConfig
	UnimplementedCoordServiceServer
}

func NewCoord() *Coord {
	return &Coord{}
}

func (coord *Coord) Start() {
	lis, _ := net.Listen("tcp", coord.CoordConfig.ListenAddr)
	defer lis.Close()
	grpcServer := grpc.NewServer()
	RegisterCoordServiceServer(grpcServer, coord)
	err := grpcServer.Serve(lis)
	util.CheckErr(err, coord.CoordConfig.ListenAddr)
}

func (coord *Coord) NodeJoin(ctx context.Context, in *NodeJoinRequest) (*NodeJoinResponse, error) {
	return &NodeJoinResponse{}, nil
}
func (coord *Coord) RequestChain(ctx context.Context, in *Chain) (*Chain, error) {
	return &Chain{}, nil
}
