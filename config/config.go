package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Torrentio struct {
		BaseURL string `yaml:"base_url"`
		Options string `yaml:"options"`
	} `yaml:"torrentio"`

	TMDB struct {
		APIKey string `yaml:"api_key"`
		CacheSize int `yaml:"cache_size"`
	} `yaml:"tmdb"`

	Regex struct {
		Season string `yaml:"season"`
		SeasonRange string `yaml:"season_range"`
		SingleEpisode string `yaml:"single_episode"`
		EpisodeRange string `yaml:"episode_range"`
		Episode string `yaml:"episode"`
		Info string `yaml:"info"`
	} `yaml:"regex"`

	Debug bool `yaml:"debug"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}