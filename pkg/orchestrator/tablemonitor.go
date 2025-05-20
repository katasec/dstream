package orchestrator

import "context"

type TableMonitor interface {
	MonitorTable(ctx context.Context) error
}
