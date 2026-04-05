package executor

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/katasec/dstream/pkg/config"
	"github.com/katasec/dstream/pkg/orasfetch"
)

// ExecuteProviderTask orchestrates independent input and output provider processes
// Data flows: Input Provider stdout → CLI → Output Provider stdin
// This is the default "run" operation.
func ExecuteProviderTask(task *config.TaskBlock) error {
	return ExecuteProviderTaskWithCommand(task, "run")
}

// ExecuteProviderTaskWithCommand orchestrates providers with a specific lifecycle command
func ExecuteProviderTaskWithCommand(task *config.TaskBlock, command string) error {
	// For lifecycle commands (init/plan/status/destroy), only the output provider runs.
	// Input providers are data readers — they don't manage infrastructure.
	if command != "run" {
		return executeOutputProviderOnly(task, command)
	}

	return executeFullPipeline(task)
}

// executeOutputProviderOnly runs only the output provider for lifecycle commands
func executeOutputProviderOnly(task *config.TaskBlock, command string) error {
	log.Info("Running lifecycle command on output provider only", "task", task.Name, "command", command)

	outputPath, err := resolveProviderPath(task.Output)
	if err != nil {
		return fmt.Errorf("resolve output provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	outputCmd := exec.CommandContext(ctx, outputPath)
	outputStdin, err := outputCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create output provider stdin pipe: %w", err)
	}
	outputStdout, err := outputCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create output provider stdout pipe: %w", err)
	}
	outputCmd.Stderr = os.Stderr

	outputConfig, err := createCommandEnvelope(func() (string, error) {
		return task.OutputConfigAsJSON()
	}, command)
	if err != nil {
		return fmt.Errorf("create output command envelope: %w", err)
	}

	log.Debug("Sending lifecycle command to output provider", "command", command, "config", outputConfig)

	if err := outputCmd.Start(); err != nil {
		return fmt.Errorf("start output provider: %w", err)
	}

	if _, err := fmt.Fprintln(outputStdin, outputConfig); err != nil {
		return fmt.Errorf("send output config: %w", err)
	}
	outputStdin.Close()

	// Wait for ready handshake, then forward remaining stdout to os.Stdout
	scanner := bufio.NewScanner(outputStdout)
	firstLine, err := waitForReady(scanner, "output-provider", 30*time.Second)
	if err != nil {
		outputCmd.Process.Kill()
		return err
	}
	// If legacy provider returned a non-handshake line, print it
	if firstLine != "" {
		fmt.Fprintln(os.Stdout, firstLine)
	}
	// Forward remaining stdout
	go func() {
		for scanner.Scan() {
			fmt.Fprintln(os.Stdout, scanner.Text())
		}
	}()

	if err := outputCmd.Wait(); err != nil {
		return fmt.Errorf("output provider failed: %w", err)
	}

	log.Info("Lifecycle command completed successfully", "task", task.Name, "command", command)
	return nil
}

// executeFullPipeline runs both input and output providers with data relay
func executeFullPipeline(task *config.TaskBlock) error {
	log.Info("Starting provider orchestration", "task", task.Name)

	// Resolve input provider binary path
	inputPath, err := resolveProviderPath(task.Input)
	if err != nil {
		return fmt.Errorf("resolve input provider: %w", err)
	}

	// Resolve output provider binary path
	outputPath, err := resolveProviderPath(task.Output)
	if err != nil {
		return fmt.Errorf("resolve output provider: %w", err)
	}

	log.Info("Provider paths resolved",
		"input_provider", inputPath,
		"output_provider", outputPath)

	// Start input and output provider processes
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Launch input provider process
	inputCmd := exec.CommandContext(ctx, inputPath)
	inputStdout, err := inputCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create input provider stdout pipe: %w", err)
	}
	inputStdin, err := inputCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create input provider stdin pipe: %w", err)
	}
	inputCmd.Stderr = os.Stderr // Forward stderr for logging

	// Launch output provider process
	outputCmd := exec.CommandContext(ctx, outputPath)
	outputStdin, err := outputCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create output provider stdin pipe: %w", err)
	}
	outputStdout, err := outputCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create output provider stdout pipe: %w", err)
	}
	outputCmd.Stderr = os.Stderr // Forward stderr for logging

	// Create command envelopes
	inputConfig, err := createCommandEnvelope(func() (string, error) {
		return task.InputConfigAsJSON()
	}, "run")
	if err != nil {
		return fmt.Errorf("create input command envelope: %w", err)
	}

	outputConfig, err := createCommandEnvelope(func() (string, error) {
		return task.OutputConfigAsJSON()
	}, "run")
	if err != nil {
		return fmt.Errorf("create output command envelope: %w", err)
	}

	log.Debug("Sending 'run' command to both providers for data processing",
		"input_config", inputConfig,
		"output_config", outputConfig)

	// Start input provider
	if err := inputCmd.Start(); err != nil {
		return fmt.Errorf("start input provider: %w", err)
	}

	// Start output provider
	if err := outputCmd.Start(); err != nil {
		inputCmd.Process.Kill() // Cleanup input process
		return fmt.Errorf("start output provider: %w", err)
	}

	// Send configuration to input provider
	if _, err := fmt.Fprintln(inputStdin, inputConfig); err != nil {
		return fmt.Errorf("send input config: %w", err)
	}
	inputStdin.Close() // Close input stdin after sending config

	// Send configuration to output provider
	if _, err := fmt.Fprintln(outputStdin, outputConfig); err != nil {
		return fmt.Errorf("send output config: %w", err)
	}

	// Wait for ready handshake from both providers before starting data relay
	inputScanner := bufio.NewScanner(inputStdout)
	outputScanner := bufio.NewScanner(outputStdout)

	inputFirstLine, err := waitForReady(inputScanner, "input-provider", 30*time.Second)
	if err != nil {
		gracefulShutdown(inputCmd, outputCmd)
		return err
	}

	outputFirstLine, err := waitForReady(outputScanner, "output-provider", 30*time.Second)
	if err != nil {
		gracefulShutdown(inputCmd, outputCmd)
		return err
	}

	// Forward output provider's non-handshake stdout to os.Stdout
	go func() {
		if outputFirstLine != "" {
			fmt.Fprintln(os.Stdout, outputFirstLine)
		}
		for outputScanner.Scan() {
			fmt.Fprintln(os.Stdout, outputScanner.Text())
		}
	}()

	// Setup graceful shutdown handling
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Goroutine to pump data from input to output
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer outputStdin.Close()

		// If input provider sent a non-handshake first line (legacy), forward it as data
		if inputFirstLine != "" {
			if _, err := fmt.Fprintln(outputStdin, inputFirstLine); err != nil {
				errChan <- fmt.Errorf("write to output provider: %w", err)
				return
			}
		}

		for inputScanner.Scan() {
			line := inputScanner.Text()
			log.Debug("Data flowing", "data", line)

			// Forward data to output provider
			if _, err := fmt.Fprintln(outputStdin, line); err != nil {
				errChan <- fmt.Errorf("write to output provider: %w", err)
				return
			}
		}

		if err := inputScanner.Err(); err != nil {
			errChan <- fmt.Errorf("read from input provider: %w", err)
		}
	}()

	// Wait for both processes to complete
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := inputCmd.Wait(); err != nil {
			errChan <- fmt.Errorf("input provider failed: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := outputCmd.Wait(); err != nil {
			errChan <- fmt.Errorf("output provider failed: %w", err)
		}
	}()

	// Wait for completion or error
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Handle graceful shutdown with timeout
	select {
	case err := <-errChan:
		if err != nil {
			log.Error("Provider execution error", "error", err.Error())
			// Graceful shutdown of both processes
			gracefulShutdown(inputCmd, outputCmd)
			return err
		}
	case <-time.After(5 * time.Minute): // Timeout for long-running tasks
		log.Info("Task timeout reached, shutting down providers")
		gracefulShutdown(inputCmd, outputCmd)
	}

	log.Info("Provider orchestration completed successfully", "task", task.Name)
	return nil
}

// providerReadySignal represents the handshake response from a provider after config validation
type providerReadySignal struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// waitForReady reads the first line from a provider's stdout and checks for the ready handshake.
// Returns nil if the provider is ready, or an error with the provider's message if it failed.
// If the provider doesn't emit a handshake (legacy), the first line is returned as non-handshake
// so the caller can decide what to do with it.
func waitForReady(scanner *bufio.Scanner, providerName string, timeout time.Duration) (firstNonHandshakeLine string, err error) {
	readyCh := make(chan string, 1)
	go func() {
		if scanner.Scan() {
			readyCh <- scanner.Text()
		} else {
			readyCh <- ""
		}
	}()

	select {
	case line := <-readyCh:
		if line == "" {
			return "", fmt.Errorf("%s: provider closed stdout without ready signal", providerName)
		}

		var signal providerReadySignal
		if err := json.Unmarshal([]byte(line), &signal); err == nil && signal.Status != "" {
			switch signal.Status {
			case "ready":
				log.Info("Provider ready", "provider", providerName)
				return "", nil
			case "error":
				return "", fmt.Errorf("%s startup failed: %s", providerName, signal.Message)
			}
		}

		// Not a handshake line — legacy provider, return the line for the caller to handle
		log.Debug("Provider did not emit handshake, treating as legacy", "provider", providerName)
		return line, nil

	case <-time.After(timeout):
		return "", fmt.Errorf("%s: timed out waiting for ready signal", providerName)
	}
}

// resolveProviderPath determines the binary path for a provider
func resolveProviderPath(block interface{}) (string, error) {
	switch b := block.(type) {
	case *config.InputBlock:
		if b.ProviderPath != "" {
			return b.ProviderPath, nil
		}
		if b.ProviderRef != "" {
			path, err := orasfetch.PullBinary(b.ProviderRef)
			if err != nil {
				return "", fmt.Errorf("pull input provider from %s: %w", b.ProviderRef, err)
			}
			return path, nil
		}
		return "", fmt.Errorf("input block must specify provider_path or provider_ref")
	
	case *config.OutputBlock:
		if b.ProviderPath != "" {
			return b.ProviderPath, nil
		}
		if b.ProviderRef != "" {
			path, err := orasfetch.PullBinary(b.ProviderRef)
			if err != nil {
				return "", fmt.Errorf("pull output provider from %s: %w", b.ProviderRef, err)
			}
			return path, nil
		}
		return "", fmt.Errorf("output block must specify provider_path or provider_ref")
	
	default:
		return "", fmt.Errorf("unsupported provider block type: %T", block)
	}
}

// gracefulShutdown attempts to gracefully terminate both provider processes
func gracefulShutdown(inputCmd, outputCmd *exec.Cmd) {
	log.Info("Initiating graceful shutdown of providers")

	// Send SIGTERM to both processes
	if inputCmd.Process != nil {
		inputCmd.Process.Signal(syscall.SIGTERM)
	}
	if outputCmd.Process != nil {
		outputCmd.Process.Signal(syscall.SIGTERM)
	}

	// Wait for graceful shutdown with timeout
	done := make(chan bool, 2)
	
	go func() {
		inputCmd.Wait()
		done <- true
	}()
	
	go func() {
		outputCmd.Wait()
		done <- true
	}()

	// Wait for both processes or timeout
	timeout := time.After(10 * time.Second)
	shutdownCount := 0

	for shutdownCount < 2 {
		select {
		case <-done:
			shutdownCount++
		case <-timeout:
			log.Warn("Graceful shutdown timeout, force killing providers")
			if inputCmd.Process != nil {
				inputCmd.Process.Kill()
			}
			if outputCmd.Process != nil {
				outputCmd.Process.Kill()
			}
			return
		}
	}

	log.Info("Graceful shutdown completed")
}

// createCommandEnvelope wraps a provider config JSON with a command header
// This implements the command envelope pattern from DESIGN_NOTES_VERB_ROUTING.md
func createCommandEnvelope(configFunc func() (string, error), command string) (string, error) {
	// Get the original config JSON
	configJSON, err := configFunc()
	if err != nil {
		return "", fmt.Errorf("get config JSON: %w", err)
	}

	// Parse the original config to ensure it's valid JSON
	var config interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return "", fmt.Errorf("parse config JSON: %w", err)
	}

	// Create the command envelope
	envelope := map[string]interface{}{
		"command": command,
		"config":  config,
	}

	// Marshal the envelope back to JSON
	envelopeJSON, err := json.Marshal(envelope)
	if err != nil {
		return "", fmt.Errorf("marshal command envelope: %w", err)
	}

	return string(envelopeJSON), nil
}
