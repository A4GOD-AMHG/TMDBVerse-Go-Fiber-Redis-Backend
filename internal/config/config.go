package config

import "os"

type Config struct {
	AccessToken string
}

func LoadConfig() *Config {
	return &Config{
		AccessToken: os.Getenv("TMDB_API_ACCESS_TOKEN"),
	}
}
