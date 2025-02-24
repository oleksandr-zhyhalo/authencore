package configuration

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type EnvConfig struct {
	IoTEndpoint        string `yaml:"iot_endpoint"`
	CAPath             string `yaml:"ca_path"`
	CertPath           string `yaml:"cert_path"`
	KeyPath            string `yaml:"key_path"`
	RoleAlias          string `yaml:"role_alias"`
	Region             string `yaml:"region"`
	CacheBufferMinutes int    `yaml:"cache_buffer_minutes"`
	MaxRetries         int    `yaml:"max_retries"`
	RetryDelayMs       int    `yaml:"retry_delay_ms"`
}

type Config struct {
	CurrentEnv string    `yaml:"current_env"`
	Dev        EnvConfig `yaml:"dev"`
	Stage      EnvConfig `yaml:"stage"`
	Prod       EnvConfig `yaml:"prod"`
}

func (c *Config) GetCurrentConfig() (*EnvConfig, error) {
	switch c.CurrentEnv {
	case "dev":
		return &c.Dev, nil
	case "stage":
		if (EnvConfig{} == c.Stage) {
			return nil, fmt.Errorf("stage environment is not configured")
		}
		return &c.Stage, nil
	case "prod":
		if (EnvConfig{} == c.Prod) {
			return nil, fmt.Errorf("prod environment is not configured")
		}
		return &c.Prod, nil
	default:
		return nil, fmt.Errorf("unknown environment: %s", c.CurrentEnv)
	}
}
func LoadConfig(configName string) (*Config, error) {
	searchPath := []string{
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "authencore"),
		filepath.Join("/etc", "authencore"),
		".",
	}

	if home, err := os.UserHomeDir(); err == nil {
		searchPath = append(searchPath, filepath.Join(home, ".config", "authencore"))
	}
	var configPath string
	for _, path := range searchPath {
		fullPath := filepath.Join(path, configName)
		if _, err := os.Stat(fullPath); err == nil {
			configPath = fullPath
			break
		}
	}
	if configPath == "" {
		return nil, fmt.Errorf("config file %s not found in search paths", configName)
	}
	return ParseConfigFile(configPath)
}

func ParseConfigFile(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %v", path, err)
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	return &config, nil
}
