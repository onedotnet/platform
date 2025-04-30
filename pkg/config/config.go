// Package config provides configuration management functionality for the platform
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	API      APIConfig
	DB       DBConfig
	Cache    CacheConfig
	Redis    RedisConfig
	Elastic  ElasticsearchConfig
	RabbitMQ RabbitMQConfig
	Auth     AuthConfig
	LogLevel string
	Logger   LoggerConfig
}

// APIConfig contains API server configuration
type APIConfig struct {
	Port    int
	Timeout time.Duration
}

// LoggerConfig contains logger configuration
type LoggerConfig struct {
	Level       string   // debug, info, warn, error, dpanic, panic, fatal
	Development bool     // if true, uses development mode for better debugging
	OutputPaths []string // list of paths to write log output to
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	JWTSecret            string        // Secret key for JWT signing
	JWTExpirationTime    time.Duration // Token expiration time
	RefreshTokenValidity time.Duration // Refresh token validity period
	GoogleClientID       string        // Google OAuth client ID
	GoogleClientSecret   string        // Google OAuth client secret
	MicrosoftClientID    string        // Microsoft Entra ID client ID
	MicrosoftTenantID    string        // Microsoft Entra ID tenant ID
	MicrosoftSecret      string        // Microsoft Entra ID client secret
	GitHubClientID       string        // GitHub OAuth client ID
	GitHubClientSecret   string        // GitHub OAuth client secret
	WeChatAppID          string        // WeChat OAuth app ID
	WeChatSecret         string        // WeChat OAuth secret
	CallbackURLBase      string        // Base URL for OAuth callbacks
}

// Setup loads the configuration from file and environment
func Setup() (*Config, error) {
	// Set up viper configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values
	setDefaults()

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// API defaults
	viper.SetDefault("api.port", 8080)
	viper.SetDefault("api.timeout", "10s")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.development", false)
	viper.SetDefault("logger.outputpaths", []string{"stdout"})

	// DB defaults
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 5432)
	viper.SetDefault("db.user", "postgres")
	viper.SetDefault("db.password", "postgres")
	viper.SetDefault("db.dbname", "platform")
	viper.SetDefault("db.sslmode", "disable")
	viper.SetDefault("db.max_open_conns", 10)
	viper.SetDefault("db.max_idle_conns", 5)
	viper.SetDefault("db.conn_max_lifetime", "1h")

	// Cache defaults
	viper.SetDefault("cache.type", "redis")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Elasticsearch defaults
	viper.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.username", "")
	viper.SetDefault("elasticsearch.password", "")
	viper.SetDefault("elasticsearch.index_name", "platform_cache")

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.host", "localhost")
	viper.SetDefault("rabbitmq.port", 5672)
	viper.SetDefault("rabbitmq.username", "guest")
	viper.SetDefault("rabbitmq.password", "guest")
	viper.SetDefault("rabbitmq.vhost", "/")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("auth.jwt_expiration_time", "24h")
	viper.SetDefault("auth.refresh_token_validity", "168h") // 7 days
	viper.SetDefault("auth.callback_url_base", "http://localhost:8080")
}
