package cdc

import "time"

// BackoffManager handles exponential backoff logic for polling intervals.
type BackoffManager struct {
	initialInterval time.Duration
	maxInterval     time.Duration
	currentInterval time.Duration
}

// NewBackoffManager initializes a new BackoffManager with the given intervals.
func NewBackoffManager(initialInterval, maxInterval time.Duration) *BackoffManager {
	return &BackoffManager{
		initialInterval: initialInterval,
		maxInterval:     maxInterval,
		currentInterval: initialInterval,
	}
}

// GetInterval returns the current interval and updates it for backoff.
func (b *BackoffManager) GetInterval() time.Duration {
	return b.currentInterval
}

// IncreaseInterval applies exponential backoff, doubling the interval up to the maxInterval.
func (b *BackoffManager) IncreaseInterval() {
	b.currentInterval *= 2
	if b.currentInterval > b.maxInterval {
		b.currentInterval = b.maxInterval
	}
}

// ResetInterval resets the interval to the initial value after detecting changes.
func (b *BackoffManager) ResetInterval() {
	b.currentInterval = b.initialInterval
}
