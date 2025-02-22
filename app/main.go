package main

import (
	"authencore/configuration"
	"authencore/core"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	if credentials, ok := core.ReadCache(); ok {
		return outputCredentials(credentials)
	}
	credentials, err := core.RetrieveCredentials(
		client,
		currentConfig.IoTEndpoint,
		currentConfig.RoleAlias,
	)
	if err != nil {
		return fmt.Errorf("failed to retrieve credentials: %v", err)
	}
	if err := core.WriteCache(credentials); err != nil {
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
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lshortfile)

	if err := run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
