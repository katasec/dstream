//go:debug x509negativeserial=1
package main

import (
	"github.com/katasec/dstream/cmd"
	"github.com/katasec/dstream/internal/logging"
)

func main() {
	logger := logging.GetLogger()
	logger.Warn("This is a warning message")
	cmd.Execute()
}
