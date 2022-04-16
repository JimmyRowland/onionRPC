package exitNode

import (
	"context"
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"

	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"google.golang.org/grpc"
)

type Node struct {
	mu         sync.Mutex
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedExitNodeServiceServer
}

func (node *Node) ExchangePublicKey(ctx context.Context, in *PublicKey) (*PublicKey, error) {
	privb, pubb := role.GetPrivateAndPublicKey()
	pubbBytes, _ := x509.MarshalPKIXPublicKey(&pubb)
	pubaParsed, _ := x509.ParsePKIXPublicKey(in.PublicKey)
	node.mu.Lock()
	defer node.mu.Unlock()
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

	serverConnection, err := net.Dial("tcp", exitLayer.ServerAddr)
	if err != nil {
		return nil, err
	}
	defer serverConnection.Close()
	serverClient := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(serverConnection))

	err = serverClient.Call(exitLayer.ServiceMethod, exitLayer.Args, &exitLayer.Res)
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
