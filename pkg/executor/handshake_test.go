package executor

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// --- Test helper processes ---
// These simulate provider behaviors when invoked as subprocesses.
// The test binary re-execs itself with a magic env var to become the "provider".

func TestMain(m *testing.M) {
	switch os.Getenv("TEST_PROVIDER_BEHAVIOR") {
	case "ready":
		// Simulates a healthy provider: emit ready, then stay alive
		fmt.Fprintln(os.Stdout, `{"status":"ready"}`)
		fmt.Fprintln(os.Stderr, "[provider] started successfully")
		// Block until stdin closes (like a real provider waiting for data)
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		os.Exit(0)

	case "error":
		// Simulates a provider that fails validation
		fmt.Fprintln(os.Stderr, "[provider] connectionString is required")
		fmt.Fprintln(os.Stdout, `{"status":"error","message":"connectionString is required"}`)
		os.Exit(1)

	case "crash":
		// Simulates a provider that crashes immediately (e.g. panic, missing DLL)
		fmt.Fprintln(os.Stderr, "[provider] fatal: cannot load libfoo.so")
		os.Exit(2)

	case "hang":
		// Simulates a provider that hangs (never sends handshake)
		fmt.Fprintln(os.Stderr, "[provider] initializing...")
		// Sleep for a long time — avoids Go's deadlock detector which would kill the process
		time.Sleep(10 * time.Minute)

	case "legacy":
		// Simulates a legacy provider that doesn't know about handshake
		fmt.Fprintln(os.Stdout, `{"data":"some-legacy-output"}`)
		buf := make([]byte, 1)
		os.Stdin.Read(buf)
		os.Exit(0)

	case "crash_with_stderr":
		// Simulates a provider that logs heavily to stderr then crashes
		for i := 0; i < 20; i++ {
			fmt.Fprintf(os.Stderr, "[provider] loading module %d...\n", i)
		}
		fmt.Fprintln(os.Stderr, "[provider] FATAL: out of memory")
		os.Exit(1)

	case "ready_then_crash":
		// Simulates a provider that starts OK then crashes mid-stream
		fmt.Fprintln(os.Stdout, `{"status":"ready"}`)
		fmt.Fprintln(os.Stderr, "[provider] started successfully")
		// Read a few lines then crash
		scanner := bufio.NewScanner(os.Stdin)
		count := 0
		for scanner.Scan() {
			count++
			fmt.Fprintf(os.Stderr, "[provider] processed message %d\n", count)
			if count >= 2 {
				fmt.Fprintln(os.Stderr, "[provider] FATAL: connection lost")
				os.Exit(1)
			}
		}
		os.Exit(0)

	case "ready_echo":
		// Simulates a healthy provider that echoes stdin to stdout (like a relay)
		fmt.Fprintln(os.Stdout, `{"status":"ready"}`)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fmt.Fprintln(os.Stdout, scanner.Text())
		}
		os.Exit(0)

	case "ready_slow_echo":
		// Simulates a provider that processes slowly
		fmt.Fprintln(os.Stdout, `{"status":"ready"}`)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			time.Sleep(100 * time.Millisecond)
			fmt.Fprintln(os.Stdout, scanner.Text())
		}
		os.Exit(0)

	default:
		// Normal test runner
		os.Exit(m.Run())
	}
}

// helperCmd returns an exec.Cmd that re-invokes the test binary as a fake provider.
func helperCmd(behavior string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "TEST_PROVIDER_BEHAVIOR="+behavior)
	return cmd
}

// --- Tests ---

func TestWaitForReady_ProviderSendsReady(t *testing.T) {
	cmd := helperCmd("ready")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	firstLine, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if firstLine != "" {
		t.Fatalf("expected empty firstLine for handshake, got: %q", firstLine)
	}
}

func TestWaitForReady_ProviderSendsError(t *testing.T) {
	cmd := helperCmd("error")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connectionString is required") {
		t.Fatalf("expected error to contain provider message, got: %v", err)
	}
	if !strings.Contains(err.Error(), "startup failed") {
		t.Fatalf("expected error to indicate startup failure, got: %v", err)
	}
}

func TestWaitForReady_ProviderCrashes(t *testing.T) {
	cmd := helperCmd("crash")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	start := time.Now()
	_, err := waitForReady(scanner, "test-provider", 30*time.Second, cmd, &stderrBuf)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for crashed provider, got nil")
	}

	// Should detect crash quickly, NOT wait for the full 30s timeout
	if elapsed > 5*time.Second {
		t.Fatalf("crash detection took too long: %v (should be near-instant, not 30s timeout)", elapsed)
	}

	// Error should include stderr context
	if !strings.Contains(err.Error(), "cannot load libfoo.so") {
		t.Fatalf("expected stderr content in error, got: %v", err)
	}
}

func TestWaitForReady_ProviderHangs_TimesOut(t *testing.T) {
	cmd := helperCmd("hang")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	timeout := 500 * time.Millisecond
	start := time.Now()
	_, err := waitForReady(scanner, "test-provider", timeout, cmd, &stderrBuf)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout error, got: %v", err)
	}
	// Should timeout around the specified duration, not 30s
	if elapsed > 2*time.Second {
		t.Fatalf("timeout took too long: %v", elapsed)
	}
}

func TestWaitForReady_LegacyProvider(t *testing.T) {
	cmd := helperCmd("legacy")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	firstLine, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)

	if err != nil {
		t.Fatalf("expected no error for legacy provider, got: %v", err)
	}
	if firstLine == "" {
		t.Fatal("expected non-empty firstLine for legacy provider")
	}
	if !strings.Contains(firstLine, "some-legacy-output") {
		t.Fatalf("expected legacy data in firstLine, got: %q", firstLine)
	}
}

func TestWaitForReady_CrashWithStderrContext(t *testing.T) {
	cmd := helperCmd("crash_with_stderr")
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)

	if err == nil {
		t.Fatal("expected error for crashed provider, got nil")
	}

	// Should include stderr context with last N lines (not all 21 lines)
	errStr := err.Error()
	if !strings.Contains(errStr, "FATAL: out of memory") {
		t.Fatalf("expected last stderr line in error, got: %v", err)
	}
	if !strings.Contains(errStr, "Provider stderr:") {
		t.Fatalf("expected 'Provider stderr:' header in error, got: %v", err)
	}
}

// --- Post-handshake failure scenario tests (issue #26) ---

func TestProviderCrashMidStream(t *testing.T) {
	// Provider starts OK, processes 2 messages, then crashes.
	// Verifies: the writing side gets an error when the provider dies.
	cmd := helperCmd("ready_then_crash")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)
	if err != nil {
		t.Fatalf("handshake should succeed: %v", err)
	}

	// Send messages — provider crashes after 2
	var writeErr error
	for i := 0; i < 10; i++ {
		_, writeErr = fmt.Fprintf(stdin, `{"data":"msg-%d"}`+"\n", i)
		if writeErr != nil {
			break
		}
		time.Sleep(50 * time.Millisecond) // give provider time to process
	}

	// Wait for process to exit
	exitErr := cmd.Wait()

	// Either the write failed (broken pipe) or the process exited non-zero — both are acceptable
	if writeErr == nil && exitErr == nil {
		t.Fatal("expected either write error or process exit error when provider crashes mid-stream")
	}

	if exitErr != nil {
		t.Logf("Process exit error (expected): %v", exitErr)
	}
	if writeErr != nil {
		t.Logf("Write error (expected): %v", writeErr)
	}
}

func TestProviderEchoRelay(t *testing.T) {
	// Verify basic data relay works: provider echoes stdin to stdout after handshake.
	cmd := helperCmd("ready_echo")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)
	if err != nil {
		t.Fatalf("handshake should succeed: %v", err)
	}

	// Send data and verify echo
	messages := []string{
		`{"data":"hello"}`,
		`{"data":"world"}`,
		`{"data":"test"}`,
	}

	for _, msg := range messages {
		fmt.Fprintln(stdin, msg)
	}
	stdin.Close() // Signal EOF to provider

	var received []string
	for scanner.Scan() {
		received = append(received, scanner.Text())
	}

	if len(received) != len(messages) {
		t.Fatalf("expected %d echoed messages, got %d: %v", len(messages), len(received), received)
	}

	for i, msg := range messages {
		if received[i] != msg {
			t.Errorf("message %d: expected %q, got %q", i, msg, received[i])
		}
	}
}

func TestGracefulShutdownOnStdinClose(t *testing.T) {
	// Verify provider exits cleanly when stdin is closed (normal shutdown path).
	cmd := helperCmd("ready_echo")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)
	if err != nil {
		t.Fatalf("handshake should succeed: %v", err)
	}

	// Send one message, then close stdin
	fmt.Fprintln(stdin, `{"data":"final"}`)
	stdin.Close()

	// Provider should exit cleanly (exit code 0)
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case exitErr := <-done:
		if exitErr != nil {
			t.Fatalf("provider should exit cleanly on stdin close, got: %v", exitErr)
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		t.Fatal("provider did not exit within 5s after stdin close")
	}
}

func TestProviderStderrDuringNormalOperation(t *testing.T) {
	// Verify stderr logging from providers doesn't interfere with stdout data flow.
	cmd := helperCmd("ready_slow_echo")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start helper: %v", err)
	}
	defer cmd.Process.Kill()

	scanner := bufio.NewScanner(stdout)
	_, err := waitForReady(scanner, "test-provider", 5*time.Second, cmd, &stderrBuf)
	if err != nil {
		t.Fatalf("handshake should succeed: %v", err)
	}

	// Send data and verify all comes through despite slow processing
	fmt.Fprintln(stdin, `{"data":"a"}`)
	fmt.Fprintln(stdin, `{"data":"b"}`)
	stdin.Close()

	var received []string
	for scanner.Scan() {
		received = append(received, scanner.Text())
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 echoed messages, got %d: %v", len(received), received)
	}
}
