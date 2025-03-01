package utils

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ExtractServerNameFromConnectionString extracts the server name from a connection string.
// It handles special cases like localhost and IP addresses by using the machine's hostname.
// This function is useful for generating unique identifiers based on server names.
func ExtractServerNameFromConnectionString(connectionString string) (string, error) {
	// Parse the connection string
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Get the server name from the host
	serverName := strings.Split(u.Host, ".")[0]
	if serverName == "" {
		return "", fmt.Errorf("server name not found in connection string")
	}

	// If the server is localhost or an IP address, use the machine's hostname
	serverName = strings.Split(serverName, ":")[0] // Remove port if present
	if strings.ToLower(serverName) == "localhost" || isIPAddress(serverName) {
		hostname, err := os.Hostname()
		if err != nil {
			return "", fmt.Errorf("failed to get hostname: %w", err)
		}
		serverName = hostname
	}

	return strings.ToLower(serverName), nil
}

// isIPAddress checks if a string is an IP address or part of one (like '127')
func isIPAddress(host string) bool {
	// Check if it's a full IP address
	if ip := net.ParseIP(host); ip != nil {
		return true
	}

	// Check if it's a partial IP (e.g. '127' from '127.0.0.1')
	if _, err := strconv.Atoi(host); err == nil {
		// It's a number, check if it's in valid IP octet range (0-255)
		if num, _ := strconv.Atoi(host); num >= 0 && num <= 255 {
			return true
		}
	}

	// Check if it has dots but isn't a full IP (e.g. '127.0')
	if strings.Contains(host, ".") {
		parts := strings.Split(host, ".")
		if len(parts) < 4 {
			// Check if all parts are valid IP octets
			valid := true
			for _, part := range parts {
				if part == "" {
					valid = false
					break
				}
				num, err := strconv.Atoi(part)
				if err != nil || num < 0 || num > 255 {
					valid = false
					break
				}
			}
			if valid {
				return true
			}
		}
	}

	return false
}
