package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"os"
	"time"
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
	if len(os.Args) != 2 {
		fmt.Println("Usage: client [config path]")
		return
	}
	configPath := os.Args[1]
	var config onionRPC.ClientConfig
	util.ReadJSONConfig(configPath, &config)
	client := onionRPC.Client{ClientConfig: config}

	client.Start(client.ClientConfig)
	operands := Operands{A: 1, B: 1}
	result := Result{}

	for {
		err := client.RpcCall(config.ServerAddr, "Server.Add", operands, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Millisecond * 200)
		err = client.RpcCall(config.ServerAddr, "Server.Multiply", Operands{A: result.Result, B: 4}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Millisecond * 200)
		err = client.RpcCall(config.ServerAddr, "Server.Subtract", Operands{A: result.Result, B: 5}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Millisecond * 200)
		err = client.RpcCall(config.ServerAddr, "Server.Divide", Operands{A: result.Result, B: 3}, &result)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(time.Millisecond * 200)
	}
}
