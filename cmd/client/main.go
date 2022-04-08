package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
)

type RandomNumber struct {
	Number int
}
type LongString struct {
	String string
	Number int
}

func main() {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("config/client_config.json", &config)
	client := onionRPC.Client{ClientConfig: config}

	client.Start(client.ClientConfig)
	returnValue := RandomNumber{Number: 111}
	err := client.RpcCall(":4322", "Server.GetRandomNumber", RandomNumber{Number: 2}, &returnValue)
	fmt.Println(err, returnValue, returnValue.Number)
	returnString := LongString{}
	err = client.RpcCall(":4322", "Server.GetLongString", LongString{}, &returnString)
	fmt.Println(err, returnString, returnString.Number, returnString.String)
	select {}
}
