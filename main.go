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

	// Get the DB
	dbConn, err := db.Connect(dstreamConfig.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Create a Database Monitor
	monitor := cdc.NewSQLServerMonitor(dbConn)

	// Initialize the checkpoints for  monitor
	monitor.InitializeCheckpointTable(dbConn)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}
	// Start Monitoring
	monitor.StartMonitoring(dbConn, *dstreamConfig)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	// Keep the application running
	select {}
}
