package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/server"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
)

func main() {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("config/client_config.json", &config)
	client := onionRPC.Client{ClientConfig: config}

	client.Start(client.ClientConfig)
	returnValue := server.RandomNumber{Number: 111}
	args := new(interface{})
	err := client.RpcCall(config.ServerAddr, "Server.GetRandomNumber", args, &returnValue)
	fmt.Println(err, returnValue.Number)
	select {}
}
