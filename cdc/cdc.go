package cdc

import (
	"database/sql"
	"log"

	"github.com/katasec/dstream/config"
)

// StartMonitoring launches monitoring for each configured table based on the SQLServerMonitor struct
func StartMonitoring(dbConn *sql.DB, cfg config.Config) {
	sqlMonitor := NewSQLServerMonitor(dbConn)
	for _, tableConfig := range cfg.Tables {
		go func(tableConfig config.TableConfig) {
			err := sqlMonitor.MonitorTable(tableConfig)
			if err != nil {
				log.Printf("Monitoring stopped for table %s due to error: %v", tableConfig.Name, err)
			}
		}(tableConfig)
	}
}
