package db

import (
	"database/sql"
	"fmt"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/katasec/dstream/sdk/logging"
)

func Connect(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	} else {
		logging.SetupBareLogger().Info("Successfully pinged database")
	}

	var log = logging.SetupBareLogger()
	log.Info("Successfully connected to database")

	return db, nil
}
