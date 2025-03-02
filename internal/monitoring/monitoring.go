package monitoring

import (
	"runtime"
	"time"

	"github.com/katasec/dstream/internal/logging"
)

var log = logging.GetLogger()

// Monitor holds the configuration for memory logging
type Monitor struct {
	interval time.Duration
}

// NewMonitor creates a new Monitoring instance and starts logging
func NewMonitor(interval time.Duration) *Monitor {
	m := &Monitor{
		interval: interval,
	}

	// Start memory logging in a separate goroutine
	go m.Start()

	return m
}

// Start logs memory usage and the number of goroutines periodically
func (m *Monitor) Start() {
	for {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		log.Info("Number of Goroutines:", "Num", runtime.NumGoroutine())
		log.Info("Total Memory Allocated: ", "Alloc", memStats.Alloc/1024/1024, "MB")
		log.Info("Total Memory System: ", "Sys", memStats.Sys/1024/1024, "MB")
		time.Sleep(m.interval)
	}
}
