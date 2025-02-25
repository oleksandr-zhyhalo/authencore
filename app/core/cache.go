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
	Credentials     Credentials `json:"credentials"`
	Expiration      time.Time   `json:"expiration"`
	CacheCreatedAt  time.Time   `json:"cache_created_at"`
	SessionDuration int         `json:"session_duration"`
}

func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "authencore", "cache.json"), nil
}
func createPathWithFile(path string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if filepath.Base(path) != "." && filepath.Base(path) != "/" {
		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
	}
	return nil
}
func ReadCache(bufferMinutes int, sessionDurationSec int) (Credentials, bool) {
	cachePath, err := getCachePath()
	if err != nil {
		log.Printf("Error getting cache path: %v", err)
		return Credentials{}, false
	}

	// Try to create the file if it doesn't exist
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		log.Printf("Cache file doesn't exist, creating at: %s", cachePath)
		if err := createPathWithFile(cachePath); err != nil {
			log.Printf("Error creating cache file: %v", err)
			return Credentials{}, false
		}

		// Write an empty cache file
		emptyCf := cacheFile{
			CacheCreatedAt:  time.Now().UTC(),
			SessionDuration: sessionDurationSec,
		}
		file, err := os.Create(cachePath)
		if err != nil {
			log.Printf("Error creating cache file after directory creation: %v", err)
			return Credentials{}, false
		}
		json.NewEncoder(file).Encode(emptyCf)
		file.Close()

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
	if cf.SessionDuration != sessionDurationSec {
		log.Printf("Session duration changed from %d to %d, invalidating cache",
			cf.SessionDuration, sessionDurationSec)
		return Credentials{}, false
	}
	sessionExpiration := cf.CacheCreatedAt.Add(time.Duration(sessionDurationSec) * time.Second)
	bufferTime := now.Add(time.Duration(bufferMinutes) * time.Minute)

	if bufferTime.Before(sessionExpiration) {
		return cf.Credentials, true
	}

	return Credentials{}, false
}

func WriteCache(creds Credentials, sessionDurationSec int) error {
	expiration, err := time.Parse(time.RFC3339, creds.Expiration)
	if err != nil {
		return fmt.Errorf("invalid expiration format (must be RFC3339): %v", err)
	}

	cf := cacheFile{
		Credentials:     creds,
		Expiration:      expiration,
		CacheCreatedAt:  time.Now().UTC(),
		SessionDuration: sessionDurationSec,
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
