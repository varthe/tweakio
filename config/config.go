package config

import (
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"tweakio/internal/logger"
)

type Config struct {
	TorrentioURL *url.URL
	TMDB struct {
		APIKey    string
		CacheSize int
	} 
	ProxyURL *url.URL
}


func LoadConfig() (*Config, error) {
	config := &Config{}

	baseRaw := getEnv("TORRENTIO_BASE_URL", "https://torrentio.strem.fun/")
	base, err := url.ParseRequestURI(baseRaw)
	if err != nil {
		return nil, err
	}
	torrentioOptions := getEnv("TORRENTIO_OPTIONS", "")
	torrentioURL, err := base.Parse(torrentioOptions)
	if err != nil {
		return nil, err
	}
	config.TorrentioURL = torrentioURL

	if key := os.Getenv("TMDB_API_KEY"); key != "" {
		config.TMDB.APIKey = key

		if val := os.Getenv("TMDB_CACHE_SIZE"); val != "" {
			if num, err := strconv.Atoi(val); err == nil {
				config.TMDB.CacheSize = num
			} else {
				return nil, errors.New("TMDB_CACHE_SIZE must be a number")
			}
		} else {
			config.TMDB.CacheSize = 1000
		}
	}

	proxy := getEnv("PROXY_URL", "")
	if proxy != "" {
		proxyURL, err := url.ParseRequestURI(proxy)
		if err != nil {
			return nil, err
		}
		config.ProxyURL = proxyURL
	}

	logger.DebugEnabled = strings.ToLower(os.Getenv("DEBUG")) == "true"
	logger.Debug("TWEAKIO", "Debug mode enabled")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

