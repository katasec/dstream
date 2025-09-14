package executor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/katasec/dstream/pkg/config"
)

// ExecuteProviderTask orchestrates independent input and output provider processes
// Data flows: Input Provider stdout → CLI → Output Provider stdin
func ExecuteProviderTask(task *config.TaskBlock) error {
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
	outputCmd.Stdout = os.Stdout // Forward stdout for final output
	outputCmd.Stderr = os.Stderr // Forward stderr for logging

	// Serialize configurations
	inputConfig, err := task.InputConfigAsJSON()
	if err != nil {
		return fmt.Errorf("serialize input config: %w", err)
	}

	outputConfig, err := task.OutputConfigAsJSON()
	if err != nil {
		return fmt.Errorf("serialize output config: %w", err)
	}

	log.Debug("Sending configurations to providers",
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

	// Setup graceful shutdown handling
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Goroutine to pump data from input to output
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer outputStdin.Close()

		scanner := bufio.NewScanner(inputStdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.Debug("Data flowing", "data", line)
			
			// Forward data to output provider
			if _, err := fmt.Fprintln(outputStdin, line); err != nil {
				errChan <- fmt.Errorf("write to output provider: %w", err)
				return
			}
		}

		if err := scanner.Err(); err != nil {
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

// resolveProviderPath determines the binary path for a provider
func resolveProviderPath(block interface{}) (string, error) {
	switch b := block.(type) {
	case *config.InputBlock:
		if b.ProviderPath != "" {
			return b.ProviderPath, nil
		}
		if b.ProviderRef != "" {
			return "", fmt.Errorf("provider_ref not yet implemented for input providers")
		}
		return "", fmt.Errorf("input block must specify provider_path or provider_ref")
	
	case *config.OutputBlock:
		if b.ProviderPath != "" {
			return b.ProviderPath, nil
		}
		if b.ProviderRef != "" {
			return "", fmt.Errorf("provider_ref not yet implemented for output providers")
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