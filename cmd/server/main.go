package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/server"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"github.com/DistributedClocks/tracing"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: client [config path]")
		return
	}
	configPath := os.Args[1]
	var config server.Config
	util.ReadJSONConfig(configPath, &config)
	_server := server.Server{Config: config}
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	err := _server.Start(_server.Config, ctracer)
	fmt.Println(err)
	select {}
}
