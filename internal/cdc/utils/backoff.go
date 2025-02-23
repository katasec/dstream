package utils

import "time"

// BackoffManager manages exponential backoff for polling intervals
type BackoffManager struct {
	currentInterval time.Duration
	maxInterval     time.Duration
	initialInterval time.Duration
}

// NewBackoffManager initializes a new BackoffManager with the given intervals.
func NewBackoffManager(initialInterval, maxInterval time.Duration) *BackoffManager {
	return &BackoffManager{
		currentInterval: initialInterval,
		maxInterval:     maxInterval,
		initialInterval: initialInterval,
	}
}

// GetInterval returns the current interval
func (b *BackoffManager) GetInterval() time.Duration {
	return b.currentInterval
}

// IncreaseInterval increases the current interval exponentially up to maxInterval
func (b *BackoffManager) IncreaseInterval() {
	newInterval := b.currentInterval * 2
	if newInterval > b.maxInterval {
		newInterval = b.maxInterval
	}
	b.currentInterval = newInterval
}

// ResetInterval resets the interval back to the initial value
func (b *BackoffManager) ResetInterval() {
	b.currentInterval = b.initialInterval
}
