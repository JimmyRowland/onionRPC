package main

import (
	"fmt"
	"github.com/DistributedClocks/tracing"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: client [config path]")
		return
	}
	configPath := os.Args[1]
	tracingServer := tracing.NewTracingServerFromFile(configPath)

	err := tracingServer.Open()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Tracing server is running")
	tracingServer.Accept() // serve requests forever
}
