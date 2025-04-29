package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConfig contains database configuration
type DBConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// DefaultDBConfig returns a default PostgreSQL configuration
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "postgres",
		DBName:       "platform",
		SSLMode:      "disable",
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		MaxLifetime:  time.Hour,
	}
}

// LoadDBConfigFromViper loads database configuration from Viper
func LoadDBConfigFromViper() *DBConfig {
	return &DBConfig{
		Host:         viper.GetString("db.host"),
		Port:         viper.GetInt("db.port"),
		User:         viper.GetString("db.user"),
		Password:     viper.GetString("db.password"),
		DBName:       viper.GetString("db.dbname"),
		SSLMode:      viper.GetString("db.sslmode"),
		MaxOpenConns: viper.GetInt("db.max_open_conns"),
		MaxIdleConns: viper.GetInt("db.max_idle_conns"),
		MaxLifetime:  viper.GetDuration("db.conn_max_lifetime"),
	}
}

// DSN returns the PostgreSQL connection string
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Connect establishes a connection to the PostgreSQL database
func Connect(config *DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	return db, nil
}
