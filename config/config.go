package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// MQTT Configuration
	MQTTBrokerURL string `env:"MQTT_BROKER_URL" envDefault:"tcp://localhost:1883"`
	MQTTTopic     string `env:"MQTT_TOPIC" envDefault:"sensors/#"`
	MQTTClientID  string `env:"MQTT_CLIENT_ID" envDefault:"go_mqtt_api_client"`
	MQTTUsername  string `env:"MQTT_USERNAME"`
	MQTTPassword  string `env:"MQTT_PASSWORD"`
	MQTTQoS       byte   `env:"MQTT_QOS" envDefault:"1"`

	// PostgreSQL Configuration
	PostgresDSN string `env:"POSTGRES_DSN" envDefault:"host=localhost user=postgres password=postgres dbname=mqtt_api port=5432 sslmode=disable"`

	// Redis Configuration
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	RedisPassword string `env:"REDIS_PASSWORD"`
	RedisDB       int    `env:"REDIS_DB" envDefault:"0"`

	// HTTP Server Configuration
	HTTPPort string `env:"HTTP_PORT" envDefault:"8080"`

	// Graceful Shutdown
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return cfg, nil
}

