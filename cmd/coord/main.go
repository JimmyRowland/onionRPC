package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/coord"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"github.com/DistributedClocks/tracing"
)

func main() {
	var config coord.CoordConfig
	util.ReadJSONConfig("config/coord_config.json", &config)
	ctracer := tracing.NewTracer(tracing.TracerConfig{
		ServerAddress:  config.TracingServerAddr,
		TracerIdentity: config.TracingIdentity,
		Secret:         config.Secret,
	})
	coord := coord.NewCoord()
	coord.Start(config.ClientAPIListenAddr, config.ServerAPIListenAddr, config.LostMsgsThresh, ctracer)
}
