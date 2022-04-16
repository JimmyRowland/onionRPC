package onionRPC

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"time"

	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/exitNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/relayNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"github.com/DistributedClocks/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Actions to be recorded by client:

type ClientStart struct {
	ClientId string
}

type RpcCall struct {
	ClientId      string
	ServiceMethod string
}

type RpcResRecvd struct {
	ClientId      string
	ServiceMethod string
	Result        interface{}
}

type OnionReq struct {
	ClientId     string
	OldExitAddr  string
	OldRelayAddr string
	OldGuardAddr string
}

type OnionResRecvd struct {
	ClientId  string
	ExitAddr  string
	RelayAddr string
	GuardAddr string
}

type SharedSecretReq struct {
	ClientId string
	NodeType string
}

type SharedSecretResRecvd struct {
	ClientId string
	NodeType string
}

// ***

type ClientConfig struct {
	ClientID          string
	TracingServerAddr string
	TracingIdentity   string
	CoordAddr         string
}

type OnionNode struct {
	RpcAddress   string
	sessionId    string
	sharedSecret cipher.Block
}

type Client struct {
	ClientConfig ClientConfig
	Guard        OnionNode
	Exit         OnionNode
	Relay        OnionNode
	Tracer       *tracing.Tracer
}

func (client *Client) Start(config ClientConfig) {
	client.ClientConfig = config
	client.Tracer = tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
	})
	trace := client.Tracer.CreateTrace()
	trace.RecordAction(ClientStart{ClientId: client.ClientConfig.ClientID})
	client.getNodes("", "", "")
	err := client.getGuardSharedSecret()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = client.getRelaySharedSecret()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = client.getExitSharedSecret()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

}

type OnionCircuitRequest struct {
	ClientID     string
	OldExitAddr  string
	OldRelayAddr string
	OldGuardAddr string
	Timestamp    time.Time
	Token        tracing.TracingToken
}

type OnionCircuitResponse struct {
	Error  error
	Guard  OnionNode
	Exit   OnionNode
	Relay  OnionNode
	Relays []OnionNode // TODO: implement multi-relay circuits
	Token  tracing.TracingToken
}

func (client *Client) getNodes(oldExitAddr, oldRelayAddr, oldGuardAddr string) {
	trace := client.Tracer.CreateTrace()
	trace.RecordAction(OnionReq{
		ClientId:     client.ClientConfig.ClientID,
		OldExitAddr:  oldExitAddr,
		OldRelayAddr: oldRelayAddr,
		OldGuardAddr: oldGuardAddr})
	// TODO: https://www.figma.com/file/kP9OXD9I8nZgCYLmx5RpY4/Untitled?node-id=34%3A44
	coordAddr := client.ClientConfig.CoordAddr
	conn, err := net.Dial("tcp", coordAddr)
	checkErr(err, "Client failed to connect to coord")

	// Send request for a new circuit
	req := OnionCircuitRequest{
		ClientID:     client.ClientConfig.ClientID,
		Timestamp:    time.Now(),
		OldExitAddr:  oldExitAddr,
		OldRelayAddr: oldRelayAddr,
		OldGuardAddr: oldGuardAddr,
		Token:        trace.GenerateToken(),
	}
	conn.Write(encode(req))

	// Receive and decode response
	recvbuf := make([]byte, 1024)
	_, err = conn.Read(recvbuf)
	checkErr(err, "Failed to read NodeJoinMessage from TCP connection\n")
	tmpbuff := bytes.NewBuffer(recvbuf)
	msg := new(OnionCircuitResponse)
	decoder := gob.NewDecoder(tmpbuff)
	decoder.Decode(msg)

	if msg.Error != nil {
		// TODO: handle case where coord does not have enough connected nodes to form circuit
		panic(msg.Error)
	}

	client.Exit = msg.Exit
	client.Guard = msg.Guard
	client.Relay = msg.Relay
	client.Tracer.ReceiveToken(msg.Token)
	trace.RecordAction(OnionResRecvd{
		ClientId:  client.ClientConfig.ClientID,
		ExitAddr:  msg.Exit.RpcAddress,
		RelayAddr: msg.Relay.RpcAddress,
		GuardAddr: msg.Guard.RpcAddress})
}

func (client *Client) getGuardSharedSecret() error {
	trace := client.Tracer.CreateTrace()
	trace.RecordAction(SharedSecretReq{ClientId: client.ClientConfig.ClientID, NodeType: "Guard"})

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
		sharedSecret := sha256.Sum256(shared.Bytes())
		client.Guard.sharedSecret, _ = aes.NewCipher(sharedSecret[:])
		client.Guard.sessionId = hex.EncodeToString(response.PublicKey)
		trace.RecordAction(SharedSecretResRecvd{ClientId: client.ClientConfig.ClientID, NodeType: "Guard"})
	default:
		fmt.Println(string(response.PublicKey))
		return errors.New("Unknown public key type")
	}
	return nil
}

func (client *Client) getRelaySharedSecret() error {
	trace := client.Tracer.CreateTrace()
	trace.RecordAction(SharedSecretReq{ClientId: client.ClientConfig.ClientID, NodeType: "Relay"})

	priva, puba := role.GetPrivateAndPublicKey()
	pubaBytes, _ := x509.MarshalPKIXPublicKey(&puba)

	conn, err := grpc.Dial(client.Relay.RpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	nodeClient := relayNode.NewRelayNodeServiceClient(conn)
	response, err := nodeClient.ExchangePublicKey(context.Background(), &relayNode.PublicKey{PublicKey: pubaBytes})
	if err != nil {
		return err
	}

	pubbParsed, _ := x509.ParsePKIXPublicKey(response.PublicKey)
	fmt.Println(string(response.PublicKey))
	fmt.Println(pubbParsed)
	switch pubb := pubbParsed.(type) {
	case *ecdsa.PublicKey:
		shared, _ := pubb.Curve.ScalarMult(pubb.X, pubb.Y, priva.D.Bytes())
		sharedSecret := sha256.Sum256(shared.Bytes())
		client.Relay.sharedSecret, _ = aes.NewCipher(sharedSecret[:])
		client.Relay.sessionId = hex.EncodeToString(response.PublicKey)
		trace.RecordAction(SharedSecretReq{ClientId: client.ClientConfig.ClientID, NodeType: "Relay"})
	default:
		fmt.Println(string(response.PublicKey))
		return errors.New("Unknown public key type")
	}
	return nil
}

func (client *Client) getExitSharedSecret() error {
	trace := client.Tracer.CreateTrace()
	trace.RecordAction(SharedSecretReq{ClientId: client.ClientConfig.ClientID, NodeType: "Exit"})

	priva, puba := role.GetPrivateAndPublicKey()
	pubaBytes, _ := x509.MarshalPKIXPublicKey(&puba)

	conn, err := grpc.Dial(client.Exit.RpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	nodeClient := exitNode.NewExitNodeServiceClient(conn)
	response, err := nodeClient.ExchangePublicKey(context.Background(), &exitNode.PublicKey{PublicKey: pubaBytes})
	if err != nil {
		return err
	}

	pubbParsed, _ := x509.ParsePKIXPublicKey(response.PublicKey)
	fmt.Println(string(response.PublicKey))
	fmt.Println(pubbParsed)
	switch pubb := pubbParsed.(type) {
	case *ecdsa.PublicKey:
		shared, _ := pubb.Curve.ScalarMult(pubb.X, pubb.Y, priva.D.Bytes())
		sharedSecret := sha256.Sum256(shared.Bytes())
		client.Exit.sharedSecret, _ = aes.NewCipher(sharedSecret[:])
		client.Exit.sessionId = hex.EncodeToString(response.PublicKey)
		trace.RecordAction(SharedSecretReq{ClientId: client.ClientConfig.ClientID, NodeType: "Exit"})
	default:
		fmt.Println(string(response.PublicKey))
		return errors.New("Unknown public key type")
	}
	return nil
}

func (client *Client) RpcCall(serverAddr string, serviceMethod string, args interface{}, res interface{}) error {
	timeouts := 0
	timeout := 1000
	done := make(chan bool)

	trace := client.Tracer.CreateTrace()
	trace.RecordAction(RpcCall{ClientId: client.ClientConfig.ClientID, ServiceMethod: serviceMethod})

	for {
		exitLayer := role.ReqExitLayer{
			Args:          args,
			ServiceMethod: serviceMethod,
			ServerAddr:    serverAddr,
			Res:           res,
		}

		relayLayer := role.ReqRelayLayer{
			ExitListenAddr: client.Exit.RpcAddress,
			ExitSessionId:  client.Exit.sessionId,
			Encrypted:      role.Encrypt(&exitLayer, client.Exit.sharedSecret),
		}
		guardLayer := role.ReqGuardLayer{
			RelayListenAddr: client.Relay.RpcAddress,
			RelaySessionId:  client.Relay.sessionId,
			Encrypted:       role.Encrypt(&relayLayer, client.Relay.sharedSecret),
		}
		plainTextLayer := guardNode.ReqEncrypted{
			SessionId: client.Guard.sessionId,
			Encrypted: role.Encrypt(&guardLayer, client.Guard.sharedSecret),
		}
		conn, err := grpc.Dial(client.Guard.RpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()
		go func() error {
			nodeClient := guardNode.NewGuardNodeServiceClient(conn)
			response, err := nodeClient.ForwardRequest(context.Background(), &plainTextLayer)
			if err != nil {
				return err
			}
			resGuardLayer := guardNode.ResEncrypted{}
			err = role.Decrypt(response.Encrypted, client.Guard.sharedSecret, &resGuardLayer)
			if err != nil {
				return err
			}
			resRelayLayer := relayNode.ResEncrypted{}
			err = role.Decrypt(resGuardLayer.Encrypted, client.Relay.sharedSecret, &resRelayLayer)
			if err != nil {
				return err
			}
			err = role.Decrypt(resRelayLayer.Encrypted, client.Exit.sharedSecret, res)
			if err != nil {
				return err
			}
			done <- true
			return nil
		}()
		select {
		case isDone := <-done:
			if isDone == true {
				trace.RecordAction(RpcResRecvd{
					ClientId:      client.ClientConfig.ClientID,
					ServiceMethod: serviceMethod,
					Result:        res})
				return nil
			}
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			timeouts++
			if timeouts >= 3 {
				client.setUpOnionChain(0)
			}
			continue
		}
		return nil
	}
}

func (client *Client) setUpOnionChain(retries int) {
	if retries >= 100 {
		panic("Could not set up onion chain")
	}
	time.Sleep(time.Millisecond * 500)
	client.getNodes(client.Exit.RpcAddress, client.Relay.RpcAddress, client.Guard.RpcAddress)
	err := client.getGuardSharedSecret()
	if err != nil {
		client.setUpOnionChain(retries + 1)
		return
	}
	err = client.getRelaySharedSecret()
	if err != nil {
		client.setUpOnionChain(retries + 1)
		return
	}
	err = client.getExitSharedSecret()
	if err != nil {
		client.setUpOnionChain(retries + 1)
		return
	}
}
