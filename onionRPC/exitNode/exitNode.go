package exitNode

import (
	"bytes"
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
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
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
	serverClient, err := DialHTTP("http://127.0.0.1:4322")
	if err != nil {
		return nil, err
	}
	defer serverClient.Close()
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

// Client represents a JSON-RPC client.
type httpClient struct {
	client *http.Client
	req    *http.Request
	resp   chan *http.Response
}

// DialHTTP creates a new RPC clients that connection to an RPC server over HTTP.
func DialHTTP(url string) (*rpc.Client, error) {
	client := new(http.Client)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return jsonrpc.NewClient(&httpClient{client, req, make(chan *http.Response)}), nil
}

// Write implements io.Writer interface.
func (c *httpClient) Write(d []byte) (n int, err error) {
	c.req.ContentLength = int64(len(d))
	c.req.Body = ioutil.NopCloser(bytes.NewReader(d))
	resp, _ := c.client.Do(c.req)
	c.resp <- resp
	return len(d), nil
}

// Read implements io.Reader interface.
func (c *httpClient) Read(p []byte) (n int, err error) {
	resp := <-c.resp
	defer resp.Body.Close()

	return resp.Body.Read(p)
}

// Close implements io.Closer interface.
func (c *httpClient) Close() error {
	c.req.Body.Close()
	return nil
}
