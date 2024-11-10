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

// NewServer initializes the server, loads the configuration, and connects to the database
func NewServer() *Server {
	// Load config file
	config, err := config.LoadConfig("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	config.CheckConfig()

	// Connect to the database
	dbConn, err := db.Connect(config.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &Server{
		config: config,
		db:     dbConn,
	}
}

// Start initializes the TableMonitoringService and begins monitoring each table in the config
func (s *Server) Start() {
	// Create a new instance of TableMonitoringService
	tableService := cdc.NewTableMonitoringService(s.db, s.config)

	// Start the TableMonitoringService
	err := tableService.StartMonitoring()
	if err != nil {
		log.Fatalf("Failed to start monitoring service: %v", err)
	}

	// Keep the application running
	select {}
}
