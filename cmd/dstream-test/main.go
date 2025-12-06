package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/katasec/testcontainers-go-presets/mssql"
)

const stateFile = ".mssql_container_id"

func main() {
	if len(os.Args) < 3 {
		usage()
		return
	}

	section := os.Args[1]
	action := os.Args[2]

	ctx := context.Background()

	switch section {
	case "mssql":
		switch action {
		case "up":
			if err := startMSSQL(ctx); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(1)
			}
			fmt.Println("MSSQL container started.")

		case "down":
			if err := stopMSSQL(ctx); err != nil {
				fmt.Println("ERROR:", err)
				os.Exit(1)
			}
			fmt.Println("MSSQL container stopped.")

		default:
			fmt.Println("Unknown action:", action)
			usage()
		}

	default:
		fmt.Println("Unknown section:", section)
		usage()
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  dstream-test mssql up")
	fmt.Println("  dstream-test mssql down")
}

// ------------------------------------------------------------
// Start MSSQL using your preset
// ------------------------------------------------------------

func startMSSQL(ctx context.Context) error {
	fmt.Println("Starting MSSQL...")

	const pw = "P@ssw0rd123"

	// Start the container using your preset
	c, err := mssql.Run(ctx, mssql.WithPassword(pw))
	if err != nil {
		return fmt.Errorf("failed to start MSSQL container: %w", err)
	}

	// Persist the container ID so we can stop it later
	id := c.GetContainerID()
	if err := os.WriteFile(stateFile, []byte(id), 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	// Show connection string for convenience
	connStr, err := mssql.ConnectionString(ctx, c, pw, "TestDB")
	if err != nil {
		return fmt.Errorf("failed to build connection string: %w", err)
	}

	fmt.Println("Connection string:")
	fmt.Println(" ", connStr)
	fmt.Println("Container ID:", id)

	return nil
}

// ------------------------------------------------------------
// Stop MSSQL using stored container ID (via docker rm -f)
// ------------------------------------------------------------

func stopMSSQL(ctx context.Context) error {
	fmt.Println("Stopping MSSQL...")

	content, err := os.ReadFile(stateFile)
	if err != nil {
		return fmt.Errorf("could not read container ID file: %w", err)
	}
	containerID := string(content)

	// Use docker CLI to force-remove the container
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", containerID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	_ = os.Remove(stateFile)

	return nil
}
