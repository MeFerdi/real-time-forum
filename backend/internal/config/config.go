package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Environment   string        `env:"ENVIRONMENT" envDefault:"development"`
	ServerAddress string        `env:"SERVER_ADDRESS" envDefault:":8080"`
	ReadTimeout   time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout  time.Duration `env:"WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout   time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`

	// Database Configuration
	DatabaseURL     string        `env:"DATABASE_URL" envDefault:"file:data.db?cache=shared&_fk=1"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"25"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`

	// Auth Configuration
	SessionSecret     string        `env:"SESSION_SECRET" envDefault:"change-me-in-production"`
	SessionTimeout    time.Duration `env:"SESSION_TIMEOUT" envDefault:"24h"`
	BcryptCost        int           `env:"BCRYPT_COST" envDefault:"12"`
	RateLimitRequests int           `env:"RATE_LIMIT_REQUESTS" envDefault:"100"`
	RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW" envDefault:"1m"`

	// WebSocket Configuration
	WebSocketReadBufferSize  int           `env:"WS_READ_BUFFER" envDefault:"1024"`
	WebSocketWriteBufferSize int           `env:"WS_WRITE_BUFFER" envDefault:"1024"`
	WebSocketPingInterval    time.Duration `env:"WS_PING_INTERVAL" envDefault:"30s"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
