package orasfetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func PullBinary(ref string) (string, error) {
	name, version, err := parseRef(ref)
	if err != nil {
		return "", err
	}

	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	binaryName := "plugin"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	cachePath := filepath.Join(os.Getenv("HOME"), ".dstream", "plugins", name, version, platform)
	pluginPath := filepath.Join(cachePath, binaryName)

	if _, err := os.Stat(pluginPath); err == nil {
		fmt.Println("[orasfetch] Using cached plugin at", pluginPath)
		return pluginPath, nil
	}

	if err := os.MkdirAll(cachePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create plugin cache dir: %w", err)
	}

	cmd := exec.Command("oras", "pull", ref, "--output", cachePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("[orasfetch] Pulling plugin from: %s → %s\n", ref, pluginPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("oras pull failed: %w", err)
	}

	// Rename the downloaded binary to "plugin" or "plugin.exe"
	sourceName := fmt.Sprintf("plugin.%s", platform)
	if runtime.GOOS == "windows" {
		sourceName += ".exe"
	}
	sourcePath := filepath.Join(cachePath, sourceName)

	if _, err := os.Stat(sourcePath); err == nil {
		if err := os.Rename(sourcePath, pluginPath); err != nil {
			return "", fmt.Errorf("failed to rename plugin binary: %w", err)
		}
		if err := os.Chmod(pluginPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to mark plugin executable: %w", err)
		}
	} else {
		return "", fmt.Errorf("expected binary %s not found in pulled artifact", sourcePath)
	}

	return pluginPath, nil
}

func parseRef(ref string) (name string, version string, err error) {
	parts := strings.Split(ref, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid plugin ref: %s", ref)
	}
	fullPath := parts[0]
	version = parts[1]

	segments := strings.Split(fullPath, "/")
	if len(segments) < 1 {
		return "", "", fmt.Errorf("invalid ref path: %s", fullPath)
	}
	name = segments[len(segments)-1]
	return name, version, nil
}
