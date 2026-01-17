package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the daemon configuration
type Config struct {
	// WatchDir is the directory to monitor for changes
	WatchDir string `yaml:"watch_dir"`

	// LogLevel controls logging verbosity (debug, info, warn, error)
	LogLevel string `yaml:"log_level"`

	// SocketPath is the IPC socket location for daemon communication
	SocketPath string `yaml:"socket_path"`

	// Git holds git monitoring configuration
	Git GitConfig `yaml:"git"`

	// Earning holds Freedom Clock weights (placeholder for Phase 5)
	Earning EarningConfig `yaml:"earning"`
}

// GitConfig holds git monitoring settings
type GitConfig struct {
	// PollInterval is how often to check for git changes
	PollInterval time.Duration `yaml:"poll_interval"`
}

// EarningConfig holds Freedom Clock earning weights (placeholder for Phase 5)
type EarningConfig struct {
	// BaseRate is the default earning rate per hour
	BaseRate float64 `yaml:"base_rate"`

	// Multipliers holds category-specific multipliers
	Multipliers map[string]float64 `yaml:"multipliers"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}

	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = "."
	}

	return &Config{
		WatchDir:   cwd,
		LogLevel:   "info",
		SocketPath: filepath.Join(homeDir, ".daemonflow", "daemonflow.sock"),
		Git: GitConfig{
			PollInterval: 5 * time.Second,
		},
		Earning: EarningConfig{
			BaseRate:    1.0,
			Multipliers: make(map[string]float64),
		},
	}
}

// LoadConfig loads configuration from a YAML file
// If path is empty, it looks for ~/.daemonflow/config.yaml
// If the file doesn't exist, it returns default configuration
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Determine config file path
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return config, nil // Return defaults if can't determine home
		}
		path = filepath.Join(homeDir, ".daemonflow", "config.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil // File doesn't exist, return defaults
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}
