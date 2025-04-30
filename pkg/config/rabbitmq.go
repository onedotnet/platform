package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// RabbitMQConfig contains RabbitMQ AMQP configuration
type RabbitMQConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	VHost    string
}

// DefaultRabbitMQConfig returns a default RabbitMQ configuration
func DefaultRabbitMQConfig() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}
}

// LoadRabbitMQConfigFromViper loads RabbitMQ configuration from Viper
func LoadRabbitMQConfigFromViper() *RabbitMQConfig {
	return &RabbitMQConfig{
		Host:     viper.GetString("rabbitmq.host"),
		Port:     viper.GetInt("rabbitmq.port"),
		Username: viper.GetString("rabbitmq.username"),
		Password: viper.GetString("rabbitmq.password"),
		VHost:    viper.GetString("rabbitmq.vhost"),
	}
}

// DSN returns the RabbitMQ connection string
func (c *RabbitMQConfig) DSN() string {
	return "amqp://" + c.Username + ":" + c.Password + "@" + c.Host + ":" +
		fmt.Sprintf("%d", c.Port) + "/" + c.VHost
}
