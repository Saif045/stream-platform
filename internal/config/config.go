package config

import "os"

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	DataDir     string

	JWTSecret  string
	HookSecret string
}

func Load() Config {
	return Config{
		HTTPAddr:    envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: requiredEnv("DATABASE_URL"),
		DataDir:     envOrDefault("DATA_DIR", "data"),

		JWTSecret:  requiredEnv("JWT_SECRET"),
		HookSecret: requiredEnv("HOOK_SECRET"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic("missing required env var: " + key)
	}
	return value
}
