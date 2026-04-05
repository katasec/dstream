package e2e

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	_ "github.com/microsoft/go-mssqldb"
)

// TestMSSQLToServiceBus is the full E2E test for the Capability Reset milestone:
//
//	MSSQL CDC -> dstream-ingester-mssql -> dstream CLI -> dstream-out-asb -> Azure Service Bus
//
// Requires: Docker (for testcontainers), ASB_CONNECTION_STRING env var, built provider binaries.
func TestMSSQLToServiceBus(t *testing.T) {
	if os.Getenv("ASB_CONNECTION_STRING") == "" {
		t.Skip("ASB_CONNECTION_STRING not set — skipping E2E test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	asbConnStr := os.Getenv("ASB_CONNECTION_STRING")

	// --- Phase 1: Spin up MSSQL with CDC ---
	t.Log("Phase 1: Starting MSSQL container with CDC...")
	tdb := StartTestDB(t, ctx)

	// --- Phase 2: Create Service Bus queues via dstream init ---
	t.Log("Phase 2: Creating Service Bus queues via dstream init...")
	queueNames := []string{"localhost-test-db-cars", "localhost-test-db-persons"}
	initQueues(t, ctx, tdb, asbConnStr)

	// Verify queues exist by creating receivers
	asbClient, err := azservicebus.NewClientFromConnectionString(asbConnStr, nil)
	if err != nil {
		t.Fatalf("failed to create ASB client: %v", err)
	}
	defer asbClient.Close(ctx)

	// Drain any leftover messages from previous runs
	for _, qName := range queueNames {
		drainQueue(t, ctx, asbClient, qName)
	}

	// --- Phase 3: Insert test data to generate CDC events ---
	t.Log("Phase 3: Inserting test data to generate CDC events...")
	tdb.InsertTestData(t, ctx)

	// Give CDC a moment to capture the changes
	time.Sleep(2 * time.Second)

	// --- Phase 4: Run the pipeline ---
	t.Log("Phase 4: Running dstream pipeline (MSSQL -> ASB)...")
	runPipeline(t, ctx, tdb, asbConnStr)

	// --- Phase 5: Verify messages on Service Bus ---
	t.Log("Phase 5: Verifying messages on Service Bus queues...")
	verifyQueue(t, ctx, asbClient, "localhost-test-db-persons", 3)
	verifyQueue(t, ctx, asbClient, "localhost-test-db-cars", 3)

	t.Log("E2E test passed: MSSQL CDC -> dstream -> Azure Service Bus")
}

// TestDatabaseSetup is a quick smoke test — just verifies testcontainers + CDC works.
func TestDatabaseSetup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	tdb := StartTestDB(t, ctx)
	tdb.InsertTestData(t, ctx)

	// Verify CDC is enabled
	var isCDCEnabled bool
	err := tdb.DB.QueryRowContext(ctx,
		"SELECT is_cdc_enabled FROM sys.databases WHERE name = 'TestDB'").Scan(&isCDCEnabled)
	if err != nil {
		t.Fatalf("failed to check CDC status: %v", err)
	}
	if !isCDCEnabled {
		t.Fatal("CDC is not enabled on TestDB")
	}

	// Verify CDC change tables exist
	for _, table := range []string{"Cars", "Persons"} {
		var exists bool
		err := tdb.DB.QueryRowContext(ctx,
			"SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END FROM cdc.change_tables WHERE source_object_id = OBJECT_ID(@p1)",
			fmt.Sprintf("dbo.%s", table)).Scan(&exists)
		if err != nil {
			t.Fatalf("failed to check CDC table for %s: %v", table, err)
		}
		if !exists {
			t.Fatalf("CDC change table for %s does not exist", table)
		}
	}

	// Verify row counts
	var personCount, carCount int
	tdb.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM dbo.Persons").Scan(&personCount)
	tdb.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM dbo.Cars").Scan(&carCount)

	if personCount != 3 {
		t.Errorf("expected 3 persons, got %d", personCount)
	}
	if carCount != 3 {
		t.Errorf("expected 3 cars, got %d", carCount)
	}

	t.Logf("Database verified: %d persons, %d cars, CDC enabled", personCount, carCount)
}

// initQueues runs the output provider directly with the "init" command to create queues.
func initQueues(t *testing.T, ctx context.Context, tdb *TestDB, asbConnStr string) {
	t.Helper()

	envelope := map[string]interface{}{
		"command": "init",
		"config": map[string]interface{}{
			"connectionString": asbConnStr,
			"resourceGroup":    "rg-dstream-dev",
			"namespace":        "sb-dstream-dev",
			"sourceHost":       "localhost",
			"sourceDatabase":   "TestDB",
			"tables":           []string{"dbo.Cars", "dbo.Persons"},
		},
	}

	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("failed to marshal init envelope: %v", err)
	}

	providerPath := findProviderBinary(t, "dstream-out-asb")
	cmd := exec.CommandContext(ctx, providerPath)
	cmd.Stdin = strings.NewReader(string(envelopeJSON) + "\n")
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("dstream-out-asb init failed: %v", err)
	}

	t.Logf("Init output: %s", string(output))
}

// runPipeline runs the MSSQL ingester piped through to the ASB output provider.
// We run the providers directly (no CLI in the middle) for test isolation,
// but the protocol is identical to what dstream CLI does.
func runPipeline(t *testing.T, ctx context.Context, tdb *TestDB, asbConnStr string) {
	t.Helper()

	// Give the pipeline a bounded time to process
	pipelineCtx, pipelineCancel := context.WithTimeout(ctx, 30*time.Second)
	defer pipelineCancel()

	// Start input provider (MSSQL ingester)
	inputEnvelope := map[string]interface{}{
		"command": "run",
		"config": map[string]interface{}{
			"db_connection_string": tdb.DStreamConnectionString(),
			"poll_interval":        "2s",
			"max_poll_interval":    "5s",
			"tables":               []string{"Cars", "Persons"},
			"lock_config": map[string]interface{}{
				"type":              "none",
				"connection_string": "",
				"container_name":    "",
			},
		},
	}
	inputJSON, _ := json.Marshal(inputEnvelope)

	inputPath := findProviderBinary(t, "dstream-ingester-mssql")
	inputCmd := exec.CommandContext(pipelineCtx, inputPath)
	inputStdin, _ := inputCmd.StdinPipe()
	inputStdout, _ := inputCmd.StdoutPipe()
	inputCmd.Stderr = os.Stderr

	// Start output provider (ASB)
	outputEnvelope := map[string]interface{}{
		"command": "run",
		"config": map[string]interface{}{
			"connectionString": asbConnStr,
			"resourceGroup":    "rg-dstream-dev",
			"namespace":        "sb-dstream-dev",
			"sourceHost":       "localhost",
			"sourceDatabase":   "TestDB",
			"tables":           []string{"dbo.Cars", "dbo.Persons"},
		},
	}
	outputJSON, _ := json.Marshal(outputEnvelope)

	outputPath := findProviderBinary(t, "dstream-out-asb")
	outputCmd := exec.CommandContext(pipelineCtx, outputPath)
	outputStdin, _ := outputCmd.StdinPipe()
	outputCmd.Stdout = os.Stdout
	outputCmd.Stderr = os.Stderr

	// Start both providers
	if err := inputCmd.Start(); err != nil {
		t.Fatalf("failed to start input provider: %v", err)
	}
	if err := outputCmd.Start(); err != nil {
		t.Fatalf("failed to start output provider: %v", err)
	}

	// Send config to both
	fmt.Fprintln(inputStdin, string(inputJSON))
	inputStdin.Close()

	fmt.Fprintln(outputStdin, string(outputJSON))

	// Relay: input stdout -> output stdin (same as dstream CLI)
	// Only count valid JSON lines as CDC events.
	messageCount := 0
	scanner := bufio.NewScanner(inputStdout)
	for scanner.Scan() {
		line := scanner.Text()

		// Only relay valid JSON (data envelopes). Skip any non-JSON lines.
		if !json.Valid([]byte(line)) {
			t.Logf("Skipping non-JSON line: %s", line[:min(len(line), 120)])
			continue
		}

		t.Logf("CDC event: %s", line[:min(len(line), 120)])
		fmt.Fprintln(outputStdin, line)
		messageCount++

		// We expect 6 messages (3 persons + 3 cars). Stop after receiving them.
		if messageCount >= 6 {
			t.Logf("Received all %d expected messages, stopping pipeline", messageCount)
			break
		}
	}

	outputStdin.Close()

	// Wait briefly for output provider to flush
	time.Sleep(2 * time.Second)

	// Kill both processes (they may still be polling)
	inputCmd.Process.Kill()
	outputCmd.Process.Kill()
	inputCmd.Wait()
	outputCmd.Wait()

	if messageCount < 6 {
		t.Logf("Warning: only received %d of 6 expected messages (CDC may need more poll cycles)", messageCount)
	}
}

// verifyQueue reads messages from a Service Bus queue and verifies the expected count.
func verifyQueue(t *testing.T, ctx context.Context, client *azservicebus.Client, queueName string, expectedCount int) {
	t.Helper()

	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		t.Fatalf("failed to create receiver for %s: %v", queueName, err)
	}
	defer receiver.Close(ctx)

	receiveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	messages, err := receiver.ReceiveMessages(receiveCtx, expectedCount+5, nil)
	if err != nil {
		t.Fatalf("failed to receive from %s: %v", queueName, err)
	}

	// Complete messages to remove from queue
	for _, msg := range messages {
		receiver.CompleteMessage(ctx, msg, nil)
		t.Logf("[%s] message: %s", queueName, string(msg.Body)[:min(len(msg.Body), 120)])
	}

	if len(messages) < expectedCount {
		t.Errorf("queue %s: expected at least %d messages, got %d", queueName, expectedCount, len(messages))
	} else {
		t.Logf("queue %s: verified %d messages", queueName, len(messages))
	}
}

// drainQueue removes any leftover messages from a queue before testing.
func drainQueue(t *testing.T, ctx context.Context, client *azservicebus.Client, queueName string) {
	t.Helper()

	receiver, err := client.NewReceiverForQueue(queueName, nil)
	if err != nil {
		return // queue might not exist yet
	}
	defer receiver.Close(ctx)

	drainCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		messages, err := receiver.ReceiveMessages(drainCtx, 100, nil)
		if err != nil || len(messages) == 0 {
			break
		}
		for _, msg := range messages {
			receiver.CompleteMessage(ctx, msg, nil)
		}
		t.Logf("Drained %d messages from %s", len(messages), queueName)
	}
}

// findProviderBinary locates a provider binary by checking common build paths.
func findProviderBinary(t *testing.T, name string) string {
	t.Helper()

	candidates := []string{
		// Local build outputs
		fmt.Sprintf("../../%s/%s", name, name),                                              // go binary in repo root
		fmt.Sprintf("../../%s/bin/Debug/net9.0/osx-arm64/%s", name, name),                   // dotnet debug
		fmt.Sprintf("../../%s/bin/Release/net9.0/osx-arm64/%s", name, name),                  // dotnet release
		fmt.Sprintf("../../../%s/%s", name, name),                                            // from test/e2e/
		fmt.Sprintf("../../../%s/bin/Debug/net9.0/osx-arm64/%s", name, name),                 // dotnet from test/e2e/
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			t.Logf("Found %s at %s", name, path)
			return path
		}
	}

	// Try PATH
	if path, err := exec.LookPath(name); err == nil {
		return path
	}

	t.Fatalf("provider binary not found: %s (tried %v)", name, candidates)
	return ""
}
