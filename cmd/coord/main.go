package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/coord"
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
	var config coord.CoordConfig
	util.ReadJSONConfig(configPath, &config)
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	coord := coord.NewCoord()
	coord.Start(config.ClientAPIListenAddr, config.ServerAPIListenAddr, config.LostMsgsThresh, ctracer)
}
