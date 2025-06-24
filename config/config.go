package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"tweakio/internal/logger"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Torrentio struct {
		BaseURL string `yaml:"base_url"`
		Options string `yaml:"options"`
	} `yaml:"torrentio"`

	TMDB struct {
		APIKey    string `yaml:"api_key"`
		CacheSize int    `yaml:"cache_size"`
	} `yaml:"tmdb"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return loadFromEnv()
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	logger.Info("TWEAKIO", "⚠️  Config file is deprecated. Use the updated Docker Compose at: https://github.com/varthe/tweakio")

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func loadFromEnv() (*Config, error) {
	var config Config

	if baseUrl := os.Getenv("TORRENTIO_BASE_URL"); baseUrl != "" {
		config.Torrentio.BaseURL = baseUrl
	} else {
		config.Torrentio.BaseURL = "https://torrentio.strem.fun/"
	}

	if options := os.Getenv("TORRENTIO_OPTIONS"); options != "" {
		config.Torrentio.Options = options
	} else {
		config.Torrentio.Options = "providers=yts,eztv,rarbg,1337x,thepiratebay,kickasstorrents,torrentgalaxy,magnetdl,horriblesubs,nyaasi,tokyotosho,anidex|sort=qualitysize|qualityfilter=scr,cam"
	}

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

	return &config, nil
}
