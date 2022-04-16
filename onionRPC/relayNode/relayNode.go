package relayNode

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
	"time"

	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedRelayNodeServiceServer
	listener net.Listener
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
	var relayLayer role.ReqRelayLayer
	err := role.Decrypt(in.Encrypted, cipher, &relayLayer)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(relayLayer.ExitListenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	nodeClient := exitNode.NewExitNodeServiceClient(conn)
	time.Sleep(time.Millisecond * 50)
	response, err := nodeClient.ForwardRequest(context.Background(), &exitNode.ReqEncrypted{
		Encrypted: relayLayer.Encrypted,
		SessionId: relayLayer.ExitSessionId,
	})
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		return nil, err
	}
	return &ResEncrypted{
		Encrypted: role.Encrypt(response, cipher),
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
	node.listener = lis
	defer lis.Close()

	grpcServer := grpc.NewServer()

	RegisterRelayNodeServiceServer(grpcServer, node)
	fmt.Println("Relay node started", node.RoleConfig.ListenAddr)
	err = grpcServer.Serve(lis)
}

func (node *Node) Close() {
	if node.listener != nil {
		node.listener.Close()
	}
}
