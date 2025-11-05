package cloudflare

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Config holds Cloudflare authentication information
type Config struct {
	APIKey    string
	AccountID string
	Email     string
}

// WranglerConfig represents the wrangler config file structure
type WranglerConfig struct {
	AccountID string `json:"account_id"`
}

// GetCredentials attempts to get Cloudflare credentials from wrangler
func GetCredentials() (*Config, error) {
	// First, try environment variables (legacy support)
	apiKey := os.Getenv("CLOUDFLARE_API_KEY")
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

	if apiKey != "" && accountID != "" {
		return &Config{
			APIKey:    apiKey,
			AccountID: accountID,
		}, nil
	}

	// Try to get credentials from wrangler
	config, err := getWranglerConfig()
	if err == nil && config.AccountID != "" {
		// Get API token from wrangler config
		token, err := getWranglerToken()
		if err == nil && token != "" {
			return &Config{
				APIKey:    token,
				AccountID: config.AccountID,
			}, nil
		}
	}

	return nil, fmt.Errorf("no Cloudflare credentials found. Run 'npx wrangler login' or set CLOUDFLARE_API_KEY and CLOUDFLARE_ACCOUNT_ID")
}

// getWranglerConfig reads the wrangler.toml file
func getWranglerConfig() (*WranglerConfig, error) {
	// Try to read wrangler.toml
	data, err := os.ReadFile("wrangler.toml")
	if err != nil {
		return nil, err
	}

	// Simple TOML parsing for account_id
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "account_id") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				accountID := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				return &WranglerConfig{AccountID: accountID}, nil
			}
		}
	}

	return nil, fmt.Errorf("account_id not found in wrangler.toml")
}

// getWranglerToken retrieves the API token from wrangler's config
func getWranglerToken() (string, error) {
	configPath, err := getWranglerConfigPath()
	if err != nil {
		return "", err
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}

	// Parse JSON config
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return "", err
	}

	// Try to extract API token
	if token, ok := config["api_token"].(string); ok && token != "" {
		return token, nil
	}

	// Try oauth_token
	if token, ok := config["oauth_token"].(string); ok && token != "" {
		return token, nil
	}

	return "", fmt.Errorf("no API token found in wrangler config")
}

// getWranglerConfigPath returns the path to wrangler's config file
func getWranglerConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("USERPROFILE")
		if configDir == "" {
			return "", fmt.Errorf("USERPROFILE not set")
		}
		configDir = filepath.Join(configDir, ".wrangler")
	case "darwin", "linux":
		configDir = os.Getenv("HOME")
		if configDir == "" {
			return "", fmt.Errorf("HOME not set")
		}
		configDir = filepath.Join(configDir, ".wrangler")
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return filepath.Join(configDir, "config", "default.toml"), nil
}

// CheckWranglerInstalled checks if wrangler is installed
func CheckWranglerInstalled() bool {
	cmd := exec.Command("npx", "wrangler", "--version")
	err := cmd.Run()
	return err == nil
}

// RunWranglerLogin prompts the user to login with wrangler
func RunWranglerLogin() error {
	fmt.Println("\n🔐 Cloudflare Authentication Required")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println("\nTo use Cloudflare AI, you need to authenticate with wrangler.")
	fmt.Println("\nPlease run the following command in your terminal:")
	fmt.Println("\n  npx wrangler login")
	fmt.Println("\nThis will:")
	fmt.Println("  1. Open your browser")
	fmt.Println("  2. Ask you to login to Cloudflare")
	fmt.Println("  3. Authorize the application")
	fmt.Println("\nAfter logging in, run this program again.")
	fmt.Println("=" + strings.Repeat("=", 50) + "\n")

	return fmt.Errorf("wrangler authentication required")
}

// GetAccountIDFromWrangler tries to get account ID using wrangler whoami
func GetAccountIDFromWrangler() (string, error) {
	cmd := exec.Command("npx", "wrangler", "whoami")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse output for account ID
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Account ID") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf("could not parse account ID from wrangler whoami")
}

// SetupWranglerConfig creates or updates wrangler.toml with account ID
func SetupWranglerConfig(accountID string) error {
	configTemplate := fmt.Sprintf(`name = "incident-ai"
main = "main.go"
compatibility_date = "2025-01-15"

# Your Cloudflare Account ID (from wrangler whoami)
account_id = "%s"

# Cloudflare Workers AI binding
[ai]
binding = "AI"
`, accountID)

	return os.WriteFile("wrangler.toml", []byte(configTemplate), 0644)
}
