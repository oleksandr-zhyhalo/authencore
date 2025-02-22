package core

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type cacheFile struct {
	Credentials Credentials `json:"credentials"`
	Expiration  time.Time   `json:"expiration"`
}

func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cacheDir := filepath.Join(homeDir, ".authencore")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "cache.json"), nil
}

func ReadCache() (Credentials, bool) {
	cachePath, err := getCachePath()
	if err != nil {
		log.Printf("Error getting cache path: %v", err)
		return Credentials{}, false
	}
	file, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, false
		}
		log.Printf("Error opening cache file: %v", err)
		return Credentials{}, false
	}
	defer file.Close()

	var cf cacheFile
	if err := json.NewDecoder(file).Decode(&cf); err != nil {
		log.Printf("Error decoding cache file: %v", err)
		return Credentials{}, false
	}
	now := time.Now().UTC()
	fmt.Printf("Expiration: %v\n", cf.Expiration)
	fmt.Printf("Time now: %v\n", now)
	if now.Add(5 * time.Minute).Before(cf.Expiration) {
		return cf.Credentials, true
	}
	return cf.Credentials, false
}

func WriteCache(creds Credentials) error {
	expiration, err := time.Parse(time.RFC3339, creds.Expiration)
	if err != nil {
		return err
	}

	cf := cacheFile{
		Credentials: creds,
		Expiration:  expiration,
	}

	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(cf); err != nil {
		return err
	}

	if err := os.Chmod(cachePath, 0600); err != nil {
		return err
	}

	return nil
}
