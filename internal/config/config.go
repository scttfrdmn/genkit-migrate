package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultSourceProvider string              `yaml:"default_source_provider"`
	DefaultTargetProvider string              `yaml:"default_target_provider"`
	Interactive           bool                `yaml:"interactive"`
	Providers             map[string]Provider `yaml:"providers"`
}

type Provider struct {
	AWS   *AWSProvider   `yaml:"aws,omitempty"`
	GCP   *GCPProvider   `yaml:"gcp,omitempty"`
	Azure *AzureProvider `yaml:"azure,omitempty"`
}

type AWSProvider struct {
	Region  string `yaml:"region"`
	Profile string `yaml:"profile"`
}

type GCPProvider struct {
	ProjectID string `yaml:"project_id"`
	Region    string `yaml:"region"`
}

type AzureProvider struct {
	SubscriptionID string `yaml:"subscription_id"`
	ResourceGroup  string `yaml:"resource_group"`
	Location       string `yaml:"location"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{
		DefaultSourceProvider: "gcp",
		DefaultTargetProvider: "aws",
		Interactive:           true,
		Providers: map[string]Provider{
			"aws": {
				AWS: &AWSProvider{
					Region:  "us-east-1",
					Profile: "default",
				},
			},
			"gcp": {
				GCP: &GCPProvider{
					Region: "us-central1",
				},
			},
		},
	}

	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func (c *Config) Save(configPath string) error {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func getDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".genkit-migrate.yaml"
	}
	return filepath.Join(home, ".genkit-migrate.yaml")
}

func (c *Config) GetAWSConfig() *AWSProvider {
	if provider, exists := c.Providers["aws"]; exists && provider.AWS != nil {
		return provider.AWS
	}
	return &AWSProvider{
		Region:  "us-east-1",
		Profile: "default",
	}
}

func (c *Config) GetGCPConfig() *GCPProvider {
	if provider, exists := c.Providers["gcp"]; exists && provider.GCP != nil {
		return provider.GCP
	}
	return &GCPProvider{
		Region: "us-central1",
	}
}
