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
	"github.com/DistributedClocks/tracing"
	"net"
	"sync"
	"time"

	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	mu         sync.Mutex
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedRelayNodeServiceServer
	listener net.Listener
	Tracer   *tracing.Tracer
}

func (node *Node) ExchangePublicKey(ctx context.Context, in *PublicKey) (*PublicKey, error) {
	trace := node.Tracer.ReceiveToken(in.Token)
	privb, pubb := role.GetPrivateAndPublicKey()
	pubbBytes, _ := x509.MarshalPKIXPublicKey(&pubb)
	pubaParsed, _ := x509.ParsePKIXPublicKey(in.PublicKey)
	node.mu.Lock()
	defer node.mu.Unlock()
	switch puba := pubaParsed.(type) {
	case *ecdsa.PublicKey:
		trace.RecordAction(role.PublicKeyRecvd{PublicKey: puba})
		shared, _ := puba.Curve.ScalarMult(puba.X, puba.Y, privb.D.Bytes())
		sharedSecret := sha256.Sum256(shared.Bytes())
		node.Role.SessionKeys[hex.EncodeToString(pubbBytes)], _ = aes.NewCipher(sharedSecret[:])
	default:
		return &PublicKey{}, errors.New("Unknown public key type")
	}
	trace.RecordAction(role.PublicKeySent{PublicKey: pubb})
	return &PublicKey{
		PublicKey: pubbBytes,
		Token:     trace.GenerateToken(),
	}, nil
}

func (node *Node) ForwardRequest(ctx context.Context, in *ReqEncrypted) (*ResEncrypted, error) {
	trace := node.Tracer.ReceiveToken(in.Token)
	cipher, ok := node.Role.SessionKeys[in.SessionId]
	if !ok {
		return nil, errors.New("Unknown client")
	}
	var relayLayer role.ReqRelayLayer
	err := role.Decrypt(in.Encrypted, cipher, &relayLayer)
	if err != nil {
		return nil, err
	}
	trace.RecordAction(relayLayer)
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
		Token:     trace.GenerateToken(),
	})
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		return nil, err
	}
	trace = node.Tracer.ReceiveToken(response.Token)
	trace.RecordAction(role.PayloadRecvd{Encrypted: response.Encrypted})
	return &ResEncrypted{
		Encrypted: role.Encrypt(response, cipher),
		Token:     trace.GenerateToken(),
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
