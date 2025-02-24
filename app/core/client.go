package core

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Credentials struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      string `json:"expiration"`
}

type credentialsResponse struct {
	Credentials Credentials `json:"credentials"`
}

func BuildClient(certFile, keyFile, caFile string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %v", err)
	}

	caData, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file: %v", err)
	}

	rootCa, err := x509.SystemCertPool()
	if err != nil {
		rootCa = x509.NewCertPool() // Fallback to empty pool if system pool is unavailable
	}

	if ok := rootCa.AppendCertsFromPEM(caData); !ok {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      rootCa,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP384,
		},
		SessionTicketsDisabled: true,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 60 * time.Second,
	}

	return client, nil
}
func RetrieveCredentials(client *http.Client, iotEndpoint, roleAlias string, maxRetries, retryDelay int) (Credentials, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		creds, err := attemptCredentialsRequest(client, iotEndpoint, roleAlias)
		if err == nil {
			return creds, nil
		}
		lastErr = err
		time.Sleep(time.Duration(retryDelay*(i+1)) * time.Millisecond)
	}
	return Credentials{}, fmt.Errorf("after %d attempts: %v", maxRetries, lastErr)
}

func attemptCredentialsRequest(client *http.Client, iotEndpoint, roleAlias string) (Credentials, error) {
	if client == nil {
		return Credentials{}, fmt.Errorf("HTTP client cannot be nil")
	}

	url := fmt.Sprintf("https://%s/role-aliases/%s/credentials",
		iotEndpoint, roleAlias)

	resp, err := client.Get(url)
	if err != nil {
		return Credentials{}, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Credentials{}, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Credentials{}, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, body)
	}

	var response credentialsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return Credentials{}, fmt.Errorf("failed to parse credentials response: %v", err)
	}

	return response.Credentials, nil
}
