package config

import "os"

type Config struct {
	ServerPort  string
	PostgresDSN string
	RedisAddr   string
	RedisPass   string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/game_server?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
