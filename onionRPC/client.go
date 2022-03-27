package onionRPC

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	ClientID          string
	TracingServerAddr string
	TracingIdentity   string
}

type OnionNode struct {
	RpcAddress   string
	sessionId    string
	sharedSecret *ecdsa.PrivateKey
}

type Client struct {
	ClientConfig ClientConfig
	Guard        OnionNode
	Exit         OnionNode
	Relay        OnionNode
}

func (client *Client) Start() {
	client.getNodes()
	err := client.getGuardSharedSecret()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func (client *Client) getNodes() {
	//TODO: https://www.figma.com/file/kP9OXD9I8nZgCYLmx5RpY4/Untitled?node-id=34%3A44
}

func (client *Client) getGuardSharedSecret() error {
	priva, puba := role.GetPrivateAndPublicKey()
	pubaBytes, _ := x509.MarshalPKIXPublicKey(&puba)

	conn, err := grpc.Dial(client.Guard.RpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	nodeClient := guardNode.NewGuardNodeServiceClient(conn)
	response, err := nodeClient.ExchangePublicKey(context.Background(), &guardNode.PublicKey{PublicKey: pubaBytes})
	if err != nil {
		return err
	}

	pubbParsed, _ := x509.ParsePKIXPublicKey(response.PublicKey)
	fmt.Println(string(response.PublicKey))
	fmt.Println(pubbParsed)
	switch pubb := pubbParsed.(type) {
	case *ecdsa.PublicKey:
		shared, _ := pubb.Curve.ScalarMult(pubb.X, pubb.Y, priva.D.Bytes())
		sharedSecretBytes := sha256.Sum256(shared.Bytes())
		client.Guard.sharedSecret, _ = x509.ParseECPrivateKey(sharedSecretBytes[:])
		client.Guard.sessionId = string(response.PublicKey)
		fmt.Println("Shared secret: ", string(sharedSecretBytes[:]), "Session id: ", client.Guard.sessionId)
	default:
		fmt.Println(string(response.PublicKey))
		return errors.New("Unknown public key type")
	}
	return nil
}
