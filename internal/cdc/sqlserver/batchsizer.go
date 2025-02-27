package sqlserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
)

// Using log from package level

const (
	defaultSampleSize      = 100
	defaultBufferFactor    = 0.2 // 20% safety margin
	defaultResampleInterval = 1 * time.Hour

	// Service Bus SKU limits
	StandardSKULimit = 256 * 1024  // 256KB
	PremiumSKULimit  = 1024 * 1024 // 1MB
)

// BatchSizer calculates and maintains optimal batch sizes for CDC reads
type BatchSizer struct {
	batchSize        atomic.Int32
	db              *sql.DB
	tableName       string
	maxMessageSize  int
	sampleSize      int
	bufferFactor    float64
	resampleInterval time.Duration

	// For monitoring/metrics
	lastSampleTime   atomic.Int64
	lastSampleSize   atomic.Int32
	lastAvgRowSize   atomic.Int32
}

// BatchSizerOption allows customizing the BatchSizer
type BatchSizerOption func(*BatchSizer)

// NewBatchSizer creates a new BatchSizer instance
func NewBatchSizer(db *sql.DB, tableName string, maxMessageSize int, opts ...BatchSizerOption) *BatchSizer {
	bs := &BatchSizer{
		db:              db,
		tableName:       tableName,
		maxMessageSize:  maxMessageSize,
		sampleSize:      defaultSampleSize,
		bufferFactor:    defaultBufferFactor,
		resampleInterval: defaultResampleInterval,
	}

	// Apply any custom options
	for _, opt := range opts {
		opt(bs)
	}

	return bs
}

// WithSampleSize sets the number of records to sample
func WithSampleSize(size int) BatchSizerOption {
	return func(bs *BatchSizer) {
		bs.sampleSize = size
	}
}

// WithBufferFactor sets the safety margin factor
func WithBufferFactor(factor float64) BatchSizerOption {
	return func(bs *BatchSizer) {
		bs.bufferFactor = factor
	}
}

// WithResampleInterval sets how often to recalculate batch size
func WithResampleInterval(interval time.Duration) BatchSizerOption {
	return func(bs *BatchSizer) {
		bs.resampleInterval = interval
	}
}

// Start begins the batch size monitoring
func (bs *BatchSizer) Start(ctx context.Context) error {
	// Do initial sampling
	if err := bs.updateBatchSize(); err != nil {
		return fmt.Errorf("initial batch size calculation failed: %w", err)
	}

	// Start background monitoring
	go bs.monitor(ctx)
	return nil
}

// GetBatchSize returns the current calculated batch size
func (bs *BatchSizer) GetBatchSize() int32 {
	return bs.batchSize.Load()
}

// Store updates the current batch size atomically
func (bs *BatchSizer) Store(size int32) {
	bs.batchSize.Store(size)
	
	log.Info("Batch size updated", 
		"table", bs.tableName,
		"newSize", size,
		"time", time.Now().Format(time.RFC3339))
}

// updateBatchSize samples records and updates the batch size
func (bs *BatchSizer) updateBatchSize() error {
	query := fmt.Sprintf(`
		SELECT TOP(%d)
			__$start_lsn,
			__$seqval,
			__$operation,
			__$update_mask,
			ID,
			Data
		FROM cdc.dbo_%s_CT
		ORDER BY __$start_lsn DESC
	`, bs.sampleSize, bs.tableName)

	rows, err := bs.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	var totalSize int64
	var count int32

	// Calculate average size of records
	for rows.Next() {
		// Scan the CDC columns
		var startLsn []byte
		var seqVal []byte
		var operation int32
		var updateMask []byte
		var id int
		var data string
		if err := rows.Scan(&startLsn, &seqVal, &operation, &updateMask, &id, &data); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Create a record map like in real usage
		record := map[string]interface{}{
			"__$start_lsn": startLsn,
			"__$seqval": seqVal,
			"__$operation": operation,
			"__$update_mask": updateMask,
			"ID": id,
			"Data": data,
		}

		// Marshal to get real size
		jsonData, err := json.Marshal(record)
		if err != nil {
			return err
		}

		totalSize += int64(len(jsonData))
		count++
	}

	if count == 0 {
		// No records to sample, use minimum batch size
		bs.Store(200) // Start with minimum expected batch size
		return nil
	}

	avgSize := float64(totalSize) / float64(count)
	// Apply buffer factor
	effectiveSize := avgSize * (1 + bs.bufferFactor)
	
	// Use fixed batch sizes based on SKU
	var newBatchSize int32
	switch bs.maxMessageSize {
	case StandardSKULimit:
		newBatchSize = 100 // Standard SKU gets 100 records per batch
	case PremiumSKULimit:
		newBatchSize = 250 // Premium SKU gets 250 records per batch
	default:
		newBatchSize = 100 // Default to Standard SKU size
	}

	// Update batch size
	bs.Store(newBatchSize)
	
	// Update metrics
	bs.lastSampleTime.Store(time.Now().Unix())
	bs.lastSampleSize.Store(count)
	bs.lastAvgRowSize.Store(int32(avgSize))

	log.Info("Sample metrics",
		"table", bs.tableName,
		"sampleSize", count,
		"avgSize", avgSize,
		"effectiveSize", effectiveSize,
		"newBatchSize", newBatchSize)

	return nil
}

// monitor periodically updates the batch size
func (bs *BatchSizer) monitor(ctx context.Context) {
	ticker := time.NewTicker(bs.resampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := bs.updateBatchSize(); err != nil {
				log.Error("Failed to update batch size",
					"table", bs.tableName,
					"error", err)
			}
		}
	}
}

// BatchSizerMetrics contains current metrics about the batch sizer
type BatchSizerMetrics struct {
	CurrentBatchSize int32
	LastSampleTime   time.Time
	LastSampleSize   int32
	AvgRowSize      int32
	MaxMessageSize   int
	BufferFactor    float64
}

// GetMetrics returns current batch sizing metrics
func (bs *BatchSizer) GetMetrics() BatchSizerMetrics {
	return BatchSizerMetrics{
		CurrentBatchSize: bs.batchSize.Load(),
		LastSampleTime:   time.Unix(bs.lastSampleTime.Load(), 0),
		LastSampleSize:   bs.lastSampleSize.Load(),
		AvgRowSize:      bs.lastAvgRowSize.Load(),
		MaxMessageSize:   bs.maxMessageSize,
		BufferFactor:    bs.bufferFactor,
	}
}
