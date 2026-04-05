package executor

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// --- Benchmark helper processes ---
// Added to TestMain in handshake_test.go:
//   "bench_input"  — emits N JSON envelopes to stdout after handshake
//   "bench_relay"  — handshake, then reads stdin and writes to stdout (the CLI relay)
//   "bench_output" — handshake, then reads stdin and counts messages

func init() {
	// Register benchmark behaviors in TestMain via env var
	// (actual TestMain switch cases are in handshake_test.go)
}

// BenchmarkPipeRelay measures the full cost of the stdin/stdout relay path:
//   input provider (stdout) → CLI scanner → output provider (stdin)
//
// This is the core data path in executeFullPipeline. We measure:
//   - Per-message latency (time to relay N messages / N)
//   - Throughput (messages/sec and MB/sec)
//
// Uses real subprocesses to capture actual OS pipe overhead.
func BenchmarkPipeRelay(b *testing.B) {
	// Generate a realistic CDC envelope (~500 bytes, typical for a row change)
	envelope := generateCDCEnvelope()
	envelopeJSON, _ := json.Marshal(envelope)
	messageSize := len(envelopeJSON)

	b.ReportAllocs()
	b.SetBytes(int64(messageSize))

	for b.Loop() {
		relayMessages(b, envelopeJSON, 1000)
	}
}

// BenchmarkPipeRelay_SmallMessage measures relay with minimal messages (~50 bytes)
func BenchmarkPipeRelay_SmallMessage(b *testing.B) {
	msg := []byte(`{"data":{"id":1},"metadata":{"table":"t"}}`)
	b.ReportAllocs()
	b.SetBytes(int64(len(msg)))

	for b.Loop() {
		relayMessages(b, msg, 1000)
	}
}

// BenchmarkPipeRelay_LargeMessage measures relay with large messages (~5KB)
func BenchmarkPipeRelay_LargeMessage(b *testing.B) {
	envelope := generateLargeCDCEnvelope()
	envelopeJSON, _ := json.Marshal(envelope)
	b.ReportAllocs()
	b.SetBytes(int64(len(envelopeJSON)))

	for b.Loop() {
		relayMessages(b, envelopeJSON, 1000)
	}
}

// BenchmarkHandshakeLatency measures the time from process start to ready signal
func BenchmarkHandshakeLatency(b *testing.B) {
	b.ReportAllocs()

	for b.Loop() {
		cmd := helperCmd("ready")
		stdout, _ := cmd.StdoutPipe()
		var stderrBuf bytes.Buffer
		cmd.Stderr = &stderrBuf

		if err := cmd.Start(); err != nil {
			b.Fatalf("failed to start: %v", err)
		}

		scanner := bufio.NewScanner(stdout)
		_, err := waitForReady(scanner, "bench", 5*time.Second, cmd, &stderrBuf)
		if err != nil {
			b.Fatalf("handshake failed: %v", err)
		}

		cmd.Process.Kill()
		cmd.Wait()
	}
}

// relayMessages simulates the CLI relay: input process writes messages,
// relay process reads from input and writes to output, output process counts.
func relayMessages(b *testing.B, message []byte, count int) {
	b.Helper()

	// Start "input" process — writes N messages after handshake
	inputCmd := helperCmd("bench_input")
	inputCmd.Env = append(inputCmd.Env,
		fmt.Sprintf("BENCH_MSG=%s", string(message)),
		fmt.Sprintf("BENCH_COUNT=%d", count),
	)
	inputStdout, _ := inputCmd.StdoutPipe()
	inputCmd.Stderr = nil

	// Start "output" process — reads stdin after handshake, counts messages
	outputCmd := helperCmd("bench_output")
	outputStdin, _ := outputCmd.StdinPipe()
	outputStdout, _ := outputCmd.StdoutPipe()
	outputCmd.Stderr = nil

	if err := inputCmd.Start(); err != nil {
		b.Fatalf("start input: %v", err)
	}
	if err := outputCmd.Start(); err != nil {
		b.Fatalf("start output: %v", err)
	}

	// Wait for handshakes
	inputScanner := bufio.NewScanner(inputStdout)
	inputScanner.Buffer(make([]byte, 64*1024), 64*1024)
	outputScanner := bufio.NewScanner(outputStdout)

	var emptyBuf bytes.Buffer
	if _, err := waitForReady(inputScanner, "input", 5*time.Second, inputCmd, &emptyBuf); err != nil {
		b.Fatalf("input handshake: %v", err)
	}
	if _, err := waitForReady(outputScanner, "output", 5*time.Second, outputCmd, &emptyBuf); err != nil {
		b.Fatalf("output handshake: %v", err)
	}

	// Relay: input stdout → output stdin (this is what the CLI does)
	relayed := 0
	for inputScanner.Scan() {
		line := inputScanner.Text()
		if _, err := fmt.Fprintln(outputStdin, line); err != nil {
			break
		}
		relayed++
	}
	outputStdin.Close()

	inputCmd.Wait()
	outputCmd.Wait()

	if relayed < count {
		b.Fatalf("only relayed %d of %d messages", relayed, count)
	}
}

// --- Envelope generators ---

func generateCDCEnvelope() map[string]interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"__$operation":    2,
			"__$start_lsn":   "0x0000002A000001B80003",
			"__$update_mask":  "0x1F",
			"PersonID":       42,
			"FirstName":      "John",
			"LastName":       "Doe",
			"Email":          "john.doe@example.com",
			"Phone":          "+1-555-0123",
			"Address":        "123 Main St, Springfield, IL 62701",
			"CreatedAt":      "2024-11-15T10:30:00Z",
			"UpdatedAt":      "2024-11-15T14:22:00Z",
		},
		"metadata": map[string]interface{}{
			"table":    "dbo.Persons",
			"host":     "localhost",
			"database": "TestDB",
			"captured": "2024-11-15T14:22:01Z",
		},
	}
}

func generateLargeCDCEnvelope() map[string]interface{} {
	// ~5KB envelope simulating a wide table with many columns
	data := make(map[string]interface{})
	data["__$operation"] = 2
	data["__$start_lsn"] = "0x0000002A000001B80003"
	for i := 0; i < 50; i++ {
		data[fmt.Sprintf("Column_%02d", i)] = fmt.Sprintf("value_%d_with_some_padding_to_make_it_realistic_%d", i, i*1000)
	}

	return map[string]interface{}{
		"data": data,
		"metadata": map[string]interface{}{
			"table":    "dbo.WideTable",
			"host":     "localhost",
			"database": "TestDB",
			"captured": "2024-11-15T14:22:01Z",
		},
	}
}
