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
	Auth     AuthConfig
	LogLevel string
}

// APIConfig contains API server configuration
type APIConfig struct {
	Port    int
	Timeout time.Duration
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
	WeChatAppID          string        // WeChat application ID
	WeChatSecret         string        // WeChat application secret
	CallbackURLBase      string        // Base URL for OAuth callbacks
}

// GetConfig returns a populated Config instance from Viper
func GetConfig() *Config {
	return &Config{
		API: APIConfig{
			Port:    viper.GetInt("api.port"),
			Timeout: viper.GetDuration("api.timeout"),
		},
		DB:       *LoadDBConfigFromViper(),
		Cache:    *LoadCacheConfigFromViper(),
		Redis:    *LoadRedisConfigFromViper(),
		Elastic:  *LoadElasticsearchConfigFromViper(),
		Auth:     *LoadAuthConfigFromViper(),
		LogLevel: viper.GetString("log.level"),
	}
}

// Setup initializes the configuration system
func Setup() (*Config, error) {
	// Set up configuration defaults
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	setDefaults()

	// Environment variables override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, using defaults
			return GetConfig(), nil
		}
		// Config file found but another error was produced
		return nil, err
	}

	return GetConfig(), nil
}

// LoadAuthConfigFromViper loads authentication configuration from Viper
func LoadAuthConfigFromViper() *AuthConfig {
	return &AuthConfig{
		JWTSecret:            viper.GetString("auth.jwt_secret"),
		JWTExpirationTime:    viper.GetDuration("auth.jwt_expiration_time"),
		RefreshTokenValidity: viper.GetDuration("auth.refresh_token_validity"),
		GoogleClientID:       viper.GetString("auth.google.client_id"),
		GoogleClientSecret:   viper.GetString("auth.google.client_secret"),
		MicrosoftClientID:    viper.GetString("auth.microsoft.client_id"),
		MicrosoftTenantID:    viper.GetString("auth.microsoft.tenant_id"),
		MicrosoftSecret:      viper.GetString("auth.microsoft.client_secret"),
		GitHubClientID:       viper.GetString("auth.github.client_id"),
		GitHubClientSecret:   viper.GetString("auth.github.client_secret"),
		WeChatAppID:          viper.GetString("auth.wechat.app_id"),
		WeChatSecret:         viper.GetString("auth.wechat.secret"),
		CallbackURLBase:      viper.GetString("auth.callback_url_base"),
	}
}

// setDefaults sets default values for all configuration options
func setDefaults() {
	// API defaults
	viper.SetDefault("api.port", 8080)
	viper.SetDefault("api.timeout", "10s")

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

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "your-secret-key-change-in-production")
	viper.SetDefault("auth.jwt_expiration_time", "24h")
	viper.SetDefault("auth.refresh_token_validity", "720h") // 30 days
	viper.SetDefault("auth.callback_url_base", "http://localhost:8080")

	// Log defaults
	viper.SetDefault("log.level", "info")
}
