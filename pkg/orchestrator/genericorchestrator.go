package orchestrator

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/locking"
	"github.com/katasec/dstream/pkg/logging"
)

// GenericTableMonitoringOrchestrator is a plugin-agnostic orchestrator
type GenericTableMonitoringOrchestrator struct {
	db              *sql.DB
	lockerFactory   *locking.LockerFactory
	tablesToMonitor []config.ResolvedTableConfig
	monitorFactory  func(table config.ResolvedTableConfig) (TableMonitor, error)
}

// NewGenericTableMonitoringOrchestrator builds a plugin-aware orchestrator
func NewGenericTableMonitoringOrchestrator(db *sql.DB, lockerFactory *locking.LockerFactory, tables []config.ResolvedTableConfig, monitorFactory func(table config.ResolvedTableConfig) (TableMonitor, error)) *GenericTableMonitoringOrchestrator {
	return &GenericTableMonitoringOrchestrator{
		db:              db,
		lockerFactory:   lockerFactory,
		tablesToMonitor: tables,
		monitorFactory:  monitorFactory,
	}
}

// Start launches table monitors concurrently and manages graceful shutdown
func (o *GenericTableMonitoringOrchestrator) Start(ctx context.Context) error {
	log := logging.GetLogger()
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Info("Starting table monitoring orchestrator")
	
	// Start monitors for each table
	for _, table := range o.tablesToMonitor {
		lockID := fmt.Sprintf("lock-%s", table.Name)
		lock, err := o.lockerFactory.CreateLocker(lockID)
		if err != nil {
			log.Error("Failed to create lock", "table", table.Name, "error", err)
			continue
		}

		if _, err := lock.AcquireLock(ctx, lockID); err != nil {
			log.Info("Skipping table (lock held by another instance)", "table", table.Name)
			continue
		}

		wg.Add(1)
		go o.startMonitor(ctx, &wg, table, lock, lockID)
	}

	// Wait for context cancellation (which will come from the framework)
	<-ctx.Done()
	log.Info("Context cancelled, shutting down orchestrator")
	
	// Wait for all monitors to complete
	wg.Wait()
	log.Info("All monitors shut down")
	return ctx.Err()
}

// startMonitor launches a single monitor instance with lock lifecycle
func (o *GenericTableMonitoringOrchestrator) startMonitor(ctx context.Context, wg *sync.WaitGroup, table config.ResolvedTableConfig, lock locking.DistributedLocker, lockID string) {
	log := logging.GetLogger()
	defer wg.Done()
	defer func() {
		if err := lock.ReleaseLock(ctx, lockID, ""); err != nil {
			log.Error("Failed to release lock", "table", table.Name, "error", err)
		} else {
			log.Info("Released lock", "table", table.Name)
		}
	}()

	monitor, err := o.monitorFactory(table)
	if err != nil {
		log.Error("Failed to create monitor", "table", table.Name, "error", err)
		return
	}

	log.Info("Started monitoring table", "table", table.Name)
	if err := monitor.MonitorTable(ctx); err != nil && err != context.Canceled {
		log.Error("Monitor failed", "table", table.Name, "error", err)
	}
}
