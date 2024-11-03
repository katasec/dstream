package main

import (
	"log"

	"github.com/katasec/dstream/cdc"
	"github.com/katasec/dstream/config"
	"github.com/katasec/dstream/db"
)

func main() {

	// Load the configuration from the HCL file
	dstreamConfig, err := config.LoadConfig("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Connect to the database
	dbConn, err := db.Connect(dstreamConfig.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Initialize the checkpoint table
	err = cdc.InitializeCheckpointTable(dbConn)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	cdc.StartMonitoring(dbConn, *dstreamConfig)

	// Keep the application running
	select {}
}
