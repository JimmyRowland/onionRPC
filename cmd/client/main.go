package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
)

type RandomNumber struct {
	number int
}

func main() {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("config/client_config.json", &config)
	client := onionRPC.Client{ClientConfig: config}

	client.Start(client.ClientConfig)
	returnValue := RandomNumber{number: 111}
	err := client.RpcCall("server", "Server.GetRandomNumber", RandomNumber{number: 2}, &returnValue)
	fmt.Println(err, returnValue)
	select {}
}
