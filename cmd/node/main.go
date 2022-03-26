package main

import (
	"context"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/guardNode"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/role"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	for i := 0; i < 4; i++ {
		listenAddr := fmt.Sprintf(":1000%d", i)
		node := onionRPC.Node{GuardNode: guardNode.Node{RoleConfig: role.RoleConfig{ListenAddr: listenAddr}}}
		go node.Start()
		time.Sleep(time.Second)

		var conn *grpc.ClientConn
		conn, err := grpc.Dial(listenAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %s", err)
		}
		defer conn.Close()

		c := guardNode.NewGuardNodeServiceClient(conn)

		response, err := c.ExchangePublicKey(context.Background(), &guardNode.PublicKey{PublicKey: "Hello From Client!"})
		if err != nil {
			log.Fatalf("Error when calling SayHello: %s", err)
		}
		log.Printf("Response from server: %s", response.PublicKey)
	}

	ch := make(chan bool)
	<-ch
}
