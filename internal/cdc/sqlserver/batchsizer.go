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
	defaultSampleSize       = 100
	defaultBufferFactor     = 0.2 // 20% safety margin
	defaultResampleInterval = 1 * time.Hour

	// Service Bus SKU limits
	StandardSKULimit = 256 * 1024  // 256KB
	PremiumSKULimit  = 1024 * 1024 // 1MB
)

// BatchSizer calculates and maintains optimal batch sizes for CDC reads
type BatchSizer struct {
	batchSize        atomic.Int32
	db               *sql.DB
	tableName        string
	maxMessageSize   int
	sampleSize       int
	bufferFactor     float64
	resampleInterval time.Duration

	// For monitoring/metrics
	lastSampleTime atomic.Int64
	lastSampleSize atomic.Int32
	lastAvgRowSize atomic.Int32
}

// BatchSizerOption allows customizing the BatchSizer
type BatchSizerOption func(*BatchSizer)

// NewBatchSizer creates a new BatchSizer instance
func NewBatchSizer(db *sql.DB, tableName string, maxMessageSize int, opts ...BatchSizerOption) *BatchSizer {
	bs := &BatchSizer{
		db:               db,
		tableName:        tableName,
		maxMessageSize:   maxMessageSize,
		sampleSize:       defaultSampleSize,
		bufferFactor:     defaultBufferFactor,
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
	size := bs.batchSize.Load()
	// Never return 0 as batch size
	if size <= 0 {
		return 100 // Default to 100 if not yet calculated
	}
	return size
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
	// First, get the column names for the CDC table
	columnsQuery := fmt.Sprintf(`
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = 'dbo_%s_CT'
		AND TABLE_SCHEMA = 'cdc'
	`, bs.tableName)

	colRows, err := bs.db.Query(columnsQuery)
	if err != nil {
		// If we can't get columns, use a simpler approach with just system columns
		log.Info("Failed to get CDC table columns, using default size estimation",
			"table", bs.tableName,
			"error", err)

		// Use a default batch size
		bs.Store(100)
		return nil
	}
	defer colRows.Close()

	// Build column list
	var columns []string
	for colRows.Next() {
		var colName string
		if err := colRows.Scan(&colName); err != nil {
			return fmt.Errorf("failed to scan column name: %v", err)
		}
		columns = append(columns, colName)
	}

	if len(columns) == 0 {
		log.Info("No columns found for CDC table, using default size estimation",
			"table", bs.tableName)
		bs.Store(100)
		return nil
	}

	// Build a query that selects all columns
	query := fmt.Sprintf(`
		SELECT TOP(%d) *
		FROM cdc.dbo_%s_CT
		ORDER BY __$start_lsn DESC, __$seqval DESC
	`, bs.sampleSize, bs.tableName)

	rows, err := bs.db.Query(query)
	if err != nil {
		log.Info("Failed to query CDC table, using default size estimation",
			"table", bs.tableName,
			"error", err)
		bs.Store(100)
		return nil
	}
	defer rows.Close()

	var totalSize int64
	var count int32

	// Get column types from the result set
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Info("Failed to get column types, using default size estimation",
			"table", bs.tableName,
			"error", err)
		bs.Store(100)
		return nil
	}

	// Calculate average size of records
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(colTypes))
		valuePtrs := make([]interface{}, len(colTypes))

		// Create pointers to each interface{} value
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the values
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Info("Failed to scan row, using default size estimation",
				"table", bs.tableName,
				"error", err)
			continue
		}

		// Create a record map like in real usage
		record := make(map[string]interface{})
		for i, col := range colTypes {
			record[col.Name()] = values[i]
		}

		// Marshal to get real size
		jsonData, err := json.Marshal(record)
		if err != nil {
			log.Info("Failed to marshal record, skipping",
				"table", bs.tableName,
				"error", err)
			continue
		}

		totalSize += int64(len(jsonData))
		count++
	}

	if count == 0 {
		// No records to sample, use minimum batch size
		bs.Store(100) // Start with minimum expected batch size
		log.Info("No records found for sampling, using default batch size", "table", bs.tableName, "defaultBatchSize", 100)
		return nil
	}

	avgSize := float64(totalSize) / float64(count)
	// Apply buffer factor
	effectiveSize := avgSize * (1 + bs.bufferFactor)

	// Calculate batch size based on message size limits
	maxRecords := int32(float64(bs.maxMessageSize) / effectiveSize)

	// Apply reasonable limits to batch size
	var newBatchSize int32
	switch {
	case maxRecords <= 0:
		newBatchSize = 50 // Minimum batch size if calculation is off
	case maxRecords > 1000:
		newBatchSize = 1000 // Cap at 1000 records per batch
	default:
		newBatchSize = maxRecords
	}

	// Ensure we have at least a minimum batch size
	if newBatchSize < 50 {
		newBatchSize = 50
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
	AvgRowSize       int32
	MaxMessageSize   int
	BufferFactor     float64
}

// GetMetrics returns current batch sizing metrics
func (bs *BatchSizer) GetMetrics() BatchSizerMetrics {
	return BatchSizerMetrics{
		CurrentBatchSize: bs.batchSize.Load(),
		LastSampleTime:   time.Unix(bs.lastSampleTime.Load(), 0),
		LastSampleSize:   bs.lastSampleSize.Load(),
		AvgRowSize:       bs.lastAvgRowSize.Load(),
		MaxMessageSize:   bs.maxMessageSize,
		BufferFactor:     bs.bufferFactor,
	}
}
