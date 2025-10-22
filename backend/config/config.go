package config

import "os"

type Config struct {
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	PostgresDSN      string
	RouterConsoleURL string
	BindHost         string
	BindPort         int
	JWTSecret        string
}

func Load() *Config {
	// Minimal loader - replace env defaults as needed
	return &Config{
		RedisAddr:        os.Getenv("REDIS_ADDR"),
		RedisPassword:    os.Getenv("REDIS_PASS"),
		RedisDB:          0,
		PostgresDSN:      os.Getenv("POSTGRES_DSN"),
		RouterConsoleURL: "http://127.0.0.1:7657",
		BindHost:         "127.0.0.1",
		BindPort:         8080,
		JWTSecret:        getenv("JWT_SECRET", "please_change_me_very_secret"),
	}
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}
