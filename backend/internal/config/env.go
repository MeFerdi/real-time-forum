package config

import "os"

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func IsProduction() bool {
	return GetEnv("ENVIRONMENT", "development") == "production"
}
