package main

import (
	"cs.ubc.ca/cpsc416/onionRPC/onionRPC/server"
	"cs.ubc.ca/cpsc416/onionRPC/util"
	"fmt"
)

func main() {
	var config server.Config
	util.ReadJSONConfig("config/server_config.json", &config)
	_server := server.Server{Config: config}

	err := _server.Start(_server.Config)
	fmt.Println(err)
	select {}
}
