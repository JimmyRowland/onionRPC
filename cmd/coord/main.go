package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/coord"
	"cs.ubc.ca/cpsc416/onionRPC/util"
)

func main() {
	var config coord.CoordConfig
	util.ReadJSONConfig("config/coord_config.json", &config)

	coord := coord.NewCoord()
	//ctracer := tracing.NewTracer(tracing.TracerConfig{
	//	ServerAddress:  config.TracingServerAddr,
	//	TracerIdentity: config.TracingIdentity,
	//	Secret:         config.Secret,
	//})
	coord.Start()
}
