package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Accounts []Account `yaml:"accounts"`
}

// Account represents a trading account configuration
type Account struct {
	Name      string `yaml:"name"`
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
	Broker    string `yaml:"broker"`
	Enabled   bool   `yaml:"enabled"`
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if len(c.Accounts) == 0 {
		return fmt.Errorf("no accounts configured")
	}

	for i, account := range c.Accounts {
		if err := account.Validate(); err != nil {
			return fmt.Errorf("account %d (%s): %w", i, account.Name, err)
		}
	}

	return nil
}

// Validate checks if an account configuration is valid
func (a *Account) Validate() error {
	if a.Name == "" {
		return fmt.Errorf("account name is required")
	}

	if a.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	if a.SecretKey == "" {
		return fmt.Errorf("secret_key is required")
	}

	if a.Broker == "" {
		return fmt.Errorf("broker is required")
	}

	// Validate broker is supported
	supportedBrokers := map[string]bool{
		"bingx": true,
	}

	if !supportedBrokers[a.Broker] {
		return fmt.Errorf("unsupported broker: %s", a.Broker)
	}

	return nil
}

// GetEnabledAccounts returns only enabled accounts
func (c *Config) GetEnabledAccounts() []Account {
	enabled := make([]Account, 0)
	for _, account := range c.Accounts {
		if account.Enabled {
			enabled = append(enabled, account)
		}
	}
	return enabled
}

// GetAccountByName returns an account by name
func (c *Config) GetAccountByName(name string) (*Account, error) {
	for _, account := range c.Accounts {
		if account.Name == name {
			return &account, nil
		}
	}
	return nil, fmt.Errorf("account not found: %s", name)
}
