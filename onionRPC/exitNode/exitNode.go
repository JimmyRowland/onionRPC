package exitNode

import (
	"context"
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"encoding/hex"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"net/rpc"
)

type Node struct {
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedExitNodeServiceServer
}

func (node *Node) ExchangePublicKey(ctx context.Context, in *PublicKey) (*PublicKey, error) {
	privb, pubb := role.GetPrivateAndPublicKey()
	pubbBytes, _ := x509.MarshalPKIXPublicKey(&pubb)
	pubaParsed, _ := x509.ParsePKIXPublicKey(in.PublicKey)
	switch puba := pubaParsed.(type) {
	case *ecdsa.PublicKey:
		shared, _ := puba.Curve.ScalarMult(puba.X, puba.Y, privb.D.Bytes())
		sharedSecret := sha256.Sum256(shared.Bytes())
		node.Role.SessionKeys[hex.EncodeToString(pubbBytes)], _ = aes.NewCipher(sharedSecret[:])
	default:
		return &PublicKey{}, errors.New("Unknown public key type")
	}
	return &PublicKey{
		PublicKey: pubbBytes,
	}, nil
}

func (node *Node) ForwardRequest(ctx context.Context, in *ReqEncrypted) (*ResEncrypted, error) {
	cipher, ok := node.Role.SessionKeys[in.SessionId]
	if !ok {
		return nil, errors.New("Unknown client")
	}
	var exitLayer role.ReqExitLayer
	err := role.Decrypt(in.Encrypted, cipher, &exitLayer)
	if err != nil {
		return nil, err
	}
	fmt.Println(exitLayer)

	serverClient, err := rpc.Dial("tcp", exitLayer.ServerAddr)
	if err != nil {
		return nil, err
	}
	defer serverClient.Close()
	err = serverClient.Call(exitLayer.ServiceMethod, &exitLayer.Args, &exitLayer.Res)
	if err != nil {
		return nil, err
	}
	return &ResEncrypted{
		Encrypted: role.Encrypt(&exitLayer.Res, cipher),
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

	RegisterExitNodeServiceServer(grpcServer, node)
	fmt.Println("Exit node started", node.RoleConfig.ListenAddr)
	err = grpcServer.Serve(lis)
	util.CheckErr(err, node.RoleConfig.ListenAddr)
}
