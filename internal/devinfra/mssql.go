package devinfra

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	mssqlContainer    = "sqlserver"
	mssqlImage        = "mcr.microsoft.com/azure-sql-edge"
	mssqlPassword     = "Passw0rd123"
	mssqlPort         = "1433"
	mssqlConfFileName = "mssql.conf"
)

func StartMSSQL(ctx context.Context) error {
	// Stop container if exists
	_, _ = runCmd(ctx, "docker", "rm", "-f", mssqlContainer)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// scripts/mssql.conf
	confPath := filepath.Join(wd, "scripts", mssqlConfFileName)
	if _, err := os.Stat(confPath); err != nil {
		return fmt.Errorf("expected %s in scripts/: %w", mssqlConfFileName, err)
	}

	args := []string{
		"run",
		"--platform", "linux/amd64",
		"-e", "ACCEPT_EULA=Y",
		"-e", "MSSQL_SA_PASSWORD=" + mssqlPassword,
		"-p", mssqlPort + ":1433",
		"--name", mssqlContainer,
		"-v", fmt.Sprintf("%s:/var/opt/mssql/mssql.conf", confPath),
		"-d",
		mssqlImage,
	}

	out, err := runCmd(ctx, "docker", args...)
	if err != nil {
		return fmt.Errorf("docker run failed: %w (output: %s)", err, out)
	}
	return nil
}

func StopMSSQL(ctx context.Context) error {
	out, err := runCmd(ctx, "docker", "rm", "-f", mssqlContainer)
	if err != nil {
		if strings.Contains(out, "No such container") {
			return nil
		}
		return fmt.Errorf("docker rm failed: %w (output: %s)", err, out)
	}
	return nil
}

func runCmd(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
