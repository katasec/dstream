package db

import (
    "database/sql"
    "fmt"
    _ "github.com/denisenkom/go-mssqldb"
)

func Connect(connectionString string) (*sql.DB, error) {
    db, err := sql.Open("sqlserver", connectionString)
    if err != nil {
        return nil, fmt.Errorf("Failed to connect to database: %v", err)
    }
    return db, nil
}
