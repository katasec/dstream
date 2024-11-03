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

func (s *Server) Start2() {
	// Load config file
	config, err := config.LoadConfig("dstream.hcl")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	s.config = config

	// Get connection string from config and connect to the DB
	dbConn, err := db.Connect(config.DBConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	s.db = dbConn
	defer dbConn.Close()

	// Create a Monitor for the DB in the config file
	monitor := cdc.NewSQLServerMonitor(dbConn)

	// Initialize db checkpoints for the monitor
	monitor.InitializeCheckpointTable(dbConn)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	// Start Monitoring
	monitor.StartMonitoring(dbConn, *config)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	// Keep the application running
	select {}
}

func (s *Server) Start() {

	// Create a Monitor for the DB in the config file
	monitor := cdc.NewSQLServerMonitor(s.db)

	// Initialize db checkpoints for the monitor
	err := monitor.InitializeCheckpointTable(s.db)
	if err != nil {
		log.Fatalf("Error initializing checkpoint table: %v", err)
	}

	// Start Monitoring
	monitor.StartMonitoring(s.db, *s.config)

	// Keep the application running
	select {}
}
