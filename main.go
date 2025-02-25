package main

import (
	"authencore/app/configuration"
	"authencore/app/core"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	defaultConfigPath        = "authencore.yml"
	credentialProcessVersion = "1"
)

type AWSCredentialProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func setupLogger() *os.File {
	logDir := "/var/log/authencore"
	logPath := filepath.Join(logDir, "error.log")

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// This will output to stderr as final fallback
		log.Fatalf("FATAL: Failed to access log file: %v (ensure /var/log/authencore exists)", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)
	return logFile
}

func run() error {
	configPath := getConfigPath()
	config, err := configuration.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %v", configPath, err)
	}

	currentConfig, err := config.GetCurrentConfig()
	if err != nil {
		return fmt.Errorf("failed to get current environment config: %v", err)
	}

	client, err := core.BuildClient(
		currentConfig.CertPath,
		currentConfig.KeyPath,
		currentConfig.CAPath,
	)
	if err != nil {
		return fmt.Errorf("failed to create TLS client: %v", err)
	}
	if credentials, ok := core.ReadCache(currentConfig.CacheBufferMinutes, currentConfig.SessionDurationSec); ok {
		return outputCredentials(credentials)
	}
	credentials, err := core.RetrieveCredentials(
		client,
		currentConfig.IoTEndpoint,
		currentConfig.RoleAlias,
		currentConfig.MaxRetries,
		currentConfig.RetryDelayMs,
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve credentials: %v", err)
	}
	if err := core.WriteCache(credentials, currentConfig.SessionDurationSec); err != nil {
		log.Printf("Failed to write cache: %v", err)
	}

	return outputCredentials(credentials)
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return defaultConfigPath
}

func outputCredentials(creds core.Credentials) error {
	if creds.AccessKeyId == "" || creds.SecretAccessKey == "" {
		return fmt.Errorf("credentials validation failed: empty AccessKeyId or SecretAccessKey")
	}
	output := AWSCredentialProcessOutput{
		Version:         1,
		AccessKeyId:     creds.AccessKeyId,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      creds.Expiration,
	}

	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(output); err != nil {
		return fmt.Errorf("failed to encode credentials: %v", err)
	}

	return nil
}

func main() {
	logFile := setupLogger()
	defer logFile.Close()

	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
