package config

import "os"

type Config struct {
	AccessToken string
	RedisURL    string
	Port        string
}

func LoadConfig() *Config {
	return &Config{
		AccessToken: os.Getenv("TMDB_API_ACCESS_TOKEN"),
		RedisURL:    getEnv("REDIS_URL", "localhost:6379"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
