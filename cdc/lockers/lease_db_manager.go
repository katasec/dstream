package lockers

import (
	"database/sql"
	"fmt"
	"log"
)

// LeaseDBManager handles lease ID persistence in the database
type LeaseDBManager struct {
	db *sql.DB
}

// NewLeaseDBManager initializes a new LeaseDBManager
func NewLeaseDBManager(db *sql.DB) *LeaseDBManager {
	manager := &LeaseDBManager{db: db}

	// Ensure the lease_ids table exists
	if err := manager.ensureTableExists(); err != nil {
		log.Fatalf("Failed to ensure lease_ids table exists: %v", err)
	}

	return manager
}

// ensureTableExists creates the lease_ids table if it does not already exist
func (m *LeaseDBManager) ensureTableExists() error {
	query := `
        IF NOT EXISTS (
            SELECT 1 
            FROM sysobjects 
            WHERE name = 'lease_ids' AND xtype = 'U'
        )
        CREATE TABLE [lease_ids] (
            [id] INT IDENTITY(1,1) PRIMARY KEY,
            [lock_name] NVARCHAR(255) NOT NULL UNIQUE,
            [lease_id] NVARCHAR(255) NOT NULL,
            [created_at] DATETIME DEFAULT GETDATE()
        )
    `
	_, err := m.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to ensure lease_ids table exists: %w", err)
	}

	log.Println("Ensured lease_ids table exists.")
	return nil
}

// StoreLeaseID stores a new lease ID in the database
func (m *LeaseDBManager) StoreLeaseID(lockName, leaseID string) error {
	query := `
        INSERT INTO [lease_ids] ([lock_name], [lease_id]) 
        VALUES (@lock_name, @lease_id)
    `
	_, err := m.db.Exec(query,
		sql.Named("lock_name", lockName),
		sql.Named("lease_id", leaseID),
	)
	if err != nil {
		return fmt.Errorf("failed to store lease ID for lock %s: %w", lockName, err)
	}

	log.Printf("Stored lease ID for lock %s successfully.", lockName)
	return nil
}

// UpdateLeaseID updates the lease ID for an existing lock in the database
func (m *LeaseDBManager) UpdateLeaseID(lockName, leaseID string) error {
	query := `
        UPDATE [lease_ids] 
        SET [lease_id] = @lease_id, [created_at] = GETDATE() 
        WHERE [lock_name] = @lock_name
    `
	_, err := m.db.Exec(query,
		sql.Named("lock_name", lockName),
		sql.Named("lease_id", leaseID),
	)
	if err != nil {
		return fmt.Errorf("failed to update lease ID for lock %s: %w", lockName, err)
	}

	log.Printf("Updated lease ID for lock %s successfully.", lockName)
	return nil
}

// GetLeaseID retrieves the lease ID for a given lock from the database
func (m *LeaseDBManager) GetLeaseID(lockName string) (string, error) {
	query := `
        SELECT [lease_id] 
        FROM [lease_ids] 
        WHERE [lock_name] = @lock_name
    `
	var leaseID string
	err := m.db.QueryRow(query,
		sql.Named("lock_name", lockName),
	).Scan(&leaseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // No lease ID found
		}
		return "", fmt.Errorf("failed to get lease ID for lock %s: %w", lockName, err)
	}

	log.Printf("Retrieved lease ID for lock %s successfully.", lockName)
	return leaseID, nil
}

// DeleteLeaseID deletes the lease ID for a given lock from the database
func (m *LeaseDBManager) DeleteLeaseID(lockName string) error {
	query := `
        DELETE FROM [lease_ids] 
        WHERE [lock_name] = @lock_name
    `
	_, err := m.db.Exec(query,
		sql.Named("lock_name", lockName),
	)
	if err != nil {
		return fmt.Errorf("failed to delete lease ID for lock %s: %w", lockName, err)
	}

	log.Printf("Deleted lease ID for lock %s successfully.", lockName)
	return nil
}
