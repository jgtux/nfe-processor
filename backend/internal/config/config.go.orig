package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB         DBConfig
	Server     ServerConfig
	RabbitMQ   RabbitMQConfig
	Quarantine QuarantineConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Port string
}

type RabbitMQConfig struct {
	URL       string
	QueueName string
}

type QuarantineConfig struct {
	TTLDays         int
	CleanupInterval int // hours between cleanup runs
}

func Load() *Config {
	_ = godotenv.Load()

	ttl, _ := strconv.Atoi(getEnv("QUARANTINE_TTL_DAYS", "30"))
	interval, _ := strconv.Atoi(getEnv("QUARANTINE_CLEANUP_INTERVAL_HOURS", "24"))

	return &Config{
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "nfe"),
			Password: getEnv("DB_PASSWORD", "nfe123"),
			Name:     getEnv("DB_NAME", "nfe_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		RabbitMQ: RabbitMQConfig{
			URL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			QueueName: getEnv("RABBITMQ_QUEUE", "nfe_queue"),
		},
		Quarantine: QuarantineConfig{
			TTLDays:         ttl,
			CleanupInterval: interval,
		},
	}
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
