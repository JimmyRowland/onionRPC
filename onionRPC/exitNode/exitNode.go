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
	"github.com/DistributedClocks/tracing"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"time"

	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"google.golang.org/grpc"
)

type ServerPayloadRecvd struct {
	Payload interface{}
}

type Node struct {
	mu         sync.Mutex
	RoleConfig role.RoleConfig
	Role       role.Role
	UnimplementedExitNodeServiceServer
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

type Request struct {
	Token tracing.TracingToken
	Args  interface{}
}

type Response struct {
	Token tracing.TracingToken
	Res   interface{}
}

func (node *Node) ForwardRequest(ctx context.Context, in *ReqEncrypted) (*ResEncrypted, error) {
	trace := node.Tracer.ReceiveToken(in.Token)
	cipher, ok := node.Role.SessionKeys[in.SessionId]
	if !ok {
		return nil, errors.New("Unknown client")
	}
	var exitLayer role.ReqExitLayer
	err := role.Decrypt(in.Encrypted, cipher, &exitLayer)
	if err != nil {
		return nil, err
	}
	trace.RecordAction(exitLayer)
	serverConnection, err := net.Dial("tcp", exitLayer.ServerAddr)
	if err != nil {
		return nil, err
	}
	defer serverConnection.Close()
	serverClient := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(serverConnection))

	time.Sleep(time.Millisecond * 50)
	req := Request{
		Token: trace.GenerateToken(),
		Args:  exitLayer.Args,
	}
	res := Response{
		Res: &exitLayer.Res,
	}
	err = serverClient.Call(exitLayer.ServiceMethod, req, &res)
	time.Sleep(time.Millisecond * 50)
	if err != nil {
		return nil, err
	}
	trace = node.Tracer.ReceiveToken(res.Token)
	trace.RecordAction(ServerPayloadRecvd{Payload: exitLayer.Res})
	return &ResEncrypted{
		Encrypted: role.Encrypt(&exitLayer.Res, cipher),
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

	RegisterExitNodeServiceServer(grpcServer, node)
	fmt.Println("Exit node started", node.RoleConfig.ListenAddr)
	err = grpcServer.Serve(lis)
}
func (node *Node) Close() {
	if node.listener != nil {
		node.listener.Close()
	}
}
