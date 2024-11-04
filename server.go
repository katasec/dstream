package main

import (
	"database/sql"
	"log"

	"github.com/katasec/dstream/cdc"
	"github.com/katasec/dstream/config"
	"github.com/katasec/dstream/db"
)

type Server struct {
	config *config.Config
	db     *sql.DB
}

func NewServer() *Server {
	// Load config file
	config, err := config.LoadConfig("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	config.CheckConfig()

	// Get connection string from config and connect to the DB
	dbConn, err := db.Connect(config.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &Server{
		config: config,
		db:     dbConn,
	}
}

func (s *Server) Start() {

	// Create a Monitor for the DB in the config file
	monitor := cdc.NewSQLServerMonitor(s.db, s.config.AzureEventHubConnectionString, s.config.EventHubName)

	// Initialize db checkpoints for the monitor
	err := monitor.InitializeCheckpointTable()
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	// Start Monitoring
	monitor.StartMonitoring(s.db, *s.config)

	// Keep the application running
	select {}
}
