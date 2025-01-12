package main

import (
	"log"

	"github.com/katasec/dstream/config"
)

func main() {

	// dStream := NewServer()
	// dStream.Start()

	// Load config file
	config, err := config.LoadConfig2("dstream2.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Println(config.Ingester.DBType)

	log.Println(config.Ingester.Tables[0].Name)
	log.Println(config.Ingester.Tables[0].PollInterval)
	log.Println(config.Ingester.Tables[0].MaxPollInterval)

	log.Println(config.Ingester.Tables[1].Name)
	log.Println(config.Ingester.Tables[1].PollInterval)
	log.Println(config.Ingester.Tables[1].MaxPollInterval)

	log.Println(config.Publisher.Output.ConnectionString)
}
