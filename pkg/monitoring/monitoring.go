package monitoring

import (
	"runtime"
	"time"

	"github.com/katasec/dstream/pkg/logging"
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

	return m
}

// Start logs memory usage and the number of goroutines periodically
func (m *Monitor) Start() {
	for {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		log.Debug("Number of Goroutines:", "Num", runtime.NumGoroutine())
		log.Debug("Total Memory Allocated: ", "Alloc in MB", memStats.Alloc/1024/1024)
		log.Debug("Total Memory System: ", "Sys in MB", memStats.Sys/1024/1024)
		time.Sleep(m.interval)
	}
}
