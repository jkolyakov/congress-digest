package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	APIKey      string
	BaseURL     string
	HTTPTimeout time.Duration
}

const defaultBaseURL = "https://api.congress.gov/v3"
const defaultTimeout = 10 * time.Second

// Load reads configuration from environment (optionally from .env in main).
func Load() (Config, error) {
	apiKey := os.Getenv("CONGRESS_API_KEY")
	if apiKey == "" {
		return Config{}, errors.New("CONGRESS_API_KEY is not set")
	}

	baseURL := os.Getenv("CONGRESS_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	timeoutStr := os.Getenv("HTTP_TIMEOUT_SECONDS")
	timeout := defaultTimeout
	if timeoutStr != "" {
		if v, err := time.ParseDuration(timeoutStr + "s"); err == nil {
			timeout = v
		}
	}

	return Config{
		APIKey:      apiKey,
		BaseURL:     baseURL,
		HTTPTimeout: timeout,
	}, nil
}