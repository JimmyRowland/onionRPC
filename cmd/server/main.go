package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/server"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
	"github.com/DistributedClocks/tracing"
)

func main() {
	var config server.Config
	util.ReadJSONConfig("config/server_config.json", &config)
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
