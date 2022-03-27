package relayNode

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type Node struct {
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedRelayNodeServiceServer
}

func (node *Node) ExchangePublicKey(ctx context.Context, in *PublicKey) (*PublicKey, error) {
	privb, pubb := role.GetPrivateAndPublicKey()
	pubbBytes, _ := x509.MarshalPKIXPublicKey(&pubb)
	pubaParsed, _ := x509.ParsePKIXPublicKey(in.PublicKey)
	switch puba := pubaParsed.(type) {
	case *ecdsa.PublicKey:
		shared, _ := puba.Curve.ScalarMult(puba.X, puba.Y, privb.D.Bytes())
		sharedSecretBytes := sha256.Sum256(shared.Bytes())
		sharedSecret, _ := x509.ParseECPrivateKey(sharedSecretBytes[:])
		node.Role.SessionKeys[string(pubbBytes)] = sharedSecret
		fmt.Println("Shared secret: ", string(sharedSecretBytes[:]), "Session id: ", string(pubbBytes))
	default:
		return &PublicKey{}, errors.New("Unknown public key type")
	}
	return &PublicKey{
		PublicKey: pubbBytes,
	}, nil
}

func (node *Node) CheckError(err error) {
	fmt.Println(err)
	if err != nil {
		node.Role.FatalError <- err
	}
}

func (node *Node) Start() {
	node.Role = role.InitRole()
	lis, err := net.Listen("tcp", node.RoleConfig.ListenAddr)
	node.CheckError(err)
	defer lis.Close()

	grpcServer := grpc.NewServer()

	RegisterRelayNodeServiceServer(grpcServer, node)
	fmt.Println("Relay node started", node.RoleConfig.ListenAddr)
	err = grpcServer.Serve(lis)
	util.CheckErr(err, node.RoleConfig.ListenAddr)
}
