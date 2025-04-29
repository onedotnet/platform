package config

import (
	"github.com/onedotnet/platform/internal/cache"
	"github.com/spf13/viper"
)

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

// LoadRedisConfigFromViper loads Redis configuration from Viper
func LoadRedisConfigFromViper() *RedisConfig {
	return &RedisConfig{
		Host:     viper.GetString("redis.host"),
		Port:     viper.GetInt("redis.port"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	}
}

// Address returns the Redis address string
func (c *RedisConfig) Address() string {
	return c.Host + ":" + viper.GetString("redis.port")
}

// ElasticsearchConfig contains Elasticsearch configuration
type ElasticsearchConfig struct {
	Addresses []string
	Username  string
	Password  string
	IndexName string
}

// DefaultElasticsearchConfig returns a default Elasticsearch configuration
func DefaultElasticsearchConfig() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "",
		Password:  "",
		IndexName: "platform_cache",
	}
}

// LoadElasticsearchConfigFromViper loads Elasticsearch configuration from Viper
func LoadElasticsearchConfigFromViper() *ElasticsearchConfig {
	return &ElasticsearchConfig{
		Addresses: viper.GetStringSlice("elasticsearch.addresses"),
		Username:  viper.GetString("elasticsearch.username"),
		Password:  viper.GetString("elasticsearch.password"),
		IndexName: viper.GetString("elasticsearch.index_name"),
	}
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	Type          string // "redis" or "elasticsearch"
	Redis         *RedisConfig
	Elasticsearch *ElasticsearchConfig
}

// DefaultCacheConfig returns a default cache configuration using Redis
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		Type:          "redis",
		Redis:         DefaultRedisConfig(),
		Elasticsearch: DefaultElasticsearchConfig(),
	}
}

// LoadCacheConfigFromViper loads cache configuration from Viper
func LoadCacheConfigFromViper() *CacheConfig {
	return &CacheConfig{
		Type:          viper.GetString("cache.type"),
		Redis:         LoadRedisConfigFromViper(),
		Elasticsearch: LoadElasticsearchConfigFromViper(),
	}
}

// NewCache creates a new cache instance based on the configuration
func NewCache(config *CacheConfig) (cache.Cache, error) {
	switch config.Type {
	case "redis":
		return cache.NewRedisCache(config.Redis.Address(), config.Redis.Password, config.Redis.DB), nil
	case "elasticsearch":
		return cache.NewElasticsearchCache(
			config.Elasticsearch.Addresses,
			config.Elasticsearch.Username,
			config.Elasticsearch.Password,
			config.Elasticsearch.IndexName,
		)
	default:
		// Default to Redis
		return cache.NewRedisCache(config.Redis.Address(), config.Redis.Password, config.Redis.DB), nil
	}
}
