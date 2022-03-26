package guardNode

import (
	"context"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type Node struct {
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedGuardNodeServiceServer
}

func (node *Node) ExchangePublicKey(ctx context.Context, in *PublicKey) (*PublicKey, error) {
	fmt.Println(in.PublicKey)
	return &PublicKey{
		PublicKey: "node.RoleConfig.PublicKey",
	}, nil
}

func (node *Node) CheckError(err error) {
	fmt.Println(err)
	if err != nil {
		node.Role.FatalError <- err
	}
}

func (node *Node) Start() error {
	node.Role = role.InitRole()
	lis, err := net.Listen("tcp", node.RoleConfig.ListenAddr)
	node.CheckError(err)
	defer lis.Close()

	grpcServer := grpc.NewServer()

	RegisterGuardNodeServiceServer(grpcServer, node)
	fmt.Println("Guard node started", node.RoleConfig.ListenAddr)
	err = grpcServer.Serve(lis)
	return err
}
