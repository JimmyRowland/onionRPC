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

type Operands struct {
	A int
	B int
}

type Result struct {
	Result int
}

func main() {
	var config onionRPC.ClientConfig
	util.ReadJSONConfig("config/client_config.json", &config)
	client := onionRPC.Client{ClientConfig: config}

	client.Start(client.ClientConfig)
	operands := Operands{A: 1, B: 1}
	result := Result{}

	for {
		err := client.RpcCall(":4322", "Server.Add", operands, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = client.RpcCall(":4322", "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = client.RpcCall(":4322", "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = client.RpcCall(":4322", "Server.Divide", Operands{A: result.Result, B: 3}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
