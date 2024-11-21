package configs

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	AWS      AWSConfig
	Email    EmailConfig
	Scraper  ScraperConfig
	Notifier NotifierConfig
}

type ServerConfig struct {
	Port            string
	Host            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	CookieSecure    bool
	TemplatesDir    string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	MaxConns int
	MaxIdle  int
	Timeout  time.Duration
}

type AWSConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
}

type EmailConfig struct {
	FromEmail     string
	TemplatePath  string
	RetryAttempts int
	RetryDelay    time.Duration
}

type ScraperConfig struct {
	Interval      time.Duration
	Timeout       time.Duration
	RetryAttempts int
}

type NotifierConfig struct {
	Interval   time.Duration
	BatchSize  int
	MaxRetries int
}

func Load() (*Config, error) {
	cfg := &Config{}

	cfg.Server = ServerConfig{
		Port:            getEnvOrDefault("SERVER_PORT", "54321"),
		Host:            getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		ReadTimeout:     getDurationOrDefault("SERVER_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:     getDurationOrDefault("SERVER_IDLE_TIMEOUT", time.Minute),
		ShutdownTimeout: getDurationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		CookieSecure:    getBoolOrDefault("COOKIE_SECURE", true),
		TemplatesDir:    getEnvOrDefault("TEMPLATES_DIR", "./ui/html"),
	}

	// Database configuration
	cfg.Database = DatabaseConfig{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "3307"),
		User:     getRequiredEnv("DB_USER"),
		Password: getRequiredEnv("DB_PASSWORD"),
		Name:     getEnvOrDefault("DB_NAME", "scraper_db"),
		MaxConns: getIntOrDefault("DB_MAX_CONNS", 25),
		MaxIdle:  getIntOrDefault("DB_MAX_IDLE", 25),
		Timeout:  getDurationOrDefault("DB_TIMEOUT", 5*time.Minute),
	}

	cfg.AWS = AWSConfig{
		Region:          getRequiredEnv("AWS_REGION"),
		AccessKeyID:     getRequiredEnv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: getRequiredEnv("AWS_SECRET_ACCESS_KEY"),
	}

	cfg.Email = EmailConfig{
		FromEmail:     getRequiredEnv("FROM_EMAIL"),
		TemplatePath:  getEnvOrDefault("EMAIL_TEMPLATE_PATH", "./templates/email"),
		RetryAttempts: getIntOrDefault("EMAIL_RETRY_ATTEMPTS", 3),
		RetryDelay:    getDurationOrDefault("EMAIL_RETRY_DELAY", time.Second*5),
	}

	cfg.Scraper = ScraperConfig{
		Interval:      getDurationOrDefault("SCRAPER_INTERVAL", 24*time.Hour),
		Timeout:       getDurationOrDefault("SCRAPER_TIMEOUT", 5*time.Minute),
		RetryAttempts: getIntOrDefault("SCRAPER_RETRY_ATTEMPTS", 3),
	}

	cfg.Notifier = NotifierConfig{
		Interval:   getDurationOrDefault("NOTIFIER_INTERVAL", 24*time.Hour),
		BatchSize:  getIntOrDefault("NOTIFIER_BATCH_SIZE", 100),
		MaxRetries: getIntOrDefault("NOTIFIER_MAX_RETRIES", 3),
	}

	return cfg, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Helper functions

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
