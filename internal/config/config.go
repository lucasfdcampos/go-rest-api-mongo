package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Workers  WorkersConfig
}

type ServerConfig struct {
	Port string
	Host string
	Mode string // e.g., "development", "production"
}

type DatabaseConfig struct {
	URI          string
	DatabaseName string
	Timeout      time.Duration
}

type KafkaConfig struct {
	Brokers               []string
	TopicUserRegistration string
	TopicUserEvents       string
	GroupID               string
}

type JWTConfig struct {
	SecretKey  string
	Expiration time.Duration
}

type WorkersConfig struct {
	PoolSize     int
	BatchSize    int
	BatchTimeout time.Duration
}

func Load() (*Config, error) {
	setDefaults()

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Host: viper.GetString("SERVER_HOST"),
			Mode: viper.GetString("SERVER_MODE"),
		},
		Database: DatabaseConfig{
			URI:          viper.GetString("MONGO_URI"),
			DatabaseName: viper.GetString("MONGO_DATABASE"),
			Timeout:      time.Duration(viper.GetInt("MONGO_TIMEOUT")) * time.Second,
		},
		Kafka: KafkaConfig{
			Brokers:               []string{viper.GetString("KAFKA_BROKERS")},
			TopicUserRegistration: viper.GetString("KAFKA_TOPIC_USER_REGISTRATION"),
			TopicUserEvents:       viper.GetString("KAFKA_TOPIC_USER_EVENTS"),
			GroupID:               viper.GetString("KAFKA_GROUP_ID"),
		},
		JWT: JWTConfig{
			SecretKey:  viper.GetString("JWT_SECRET"),
			Expiration: time.Duration(viper.GetInt("JWT_EXPIRATION_HOURS")) * time.Hour,
		},
		Workers: WorkersConfig{
			PoolSize:     viper.GetInt("WORKER_POOL_SIZE"),
			BatchSize:    viper.GetInt("BATCH_SIZE"),
			BatchTimeout: time.Duration(viper.GetInt("BATCH_TIMEOUT_SECONDS")) * time.Second,
		},
	}

	return config, nil
}

func setDefaults() {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_MODE", "production")

	viper.SetDefault("MONGO_DB_NAME", "appdb")
	viper.SetDefault("MONGO_TIMEOUT", 10*time.Second)

	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("KAFKA_TOPIC_USER_REGISTRATION", "user-registration-topic")
	viper.SetDefault("KAFKA_TOPIC_USER_EVENTS", "user-events-topic")
	viper.SetDefault("KAFKA_GROUP_ID", "app-group")

	viper.SetDefault("JWT_SECRET_KEY", "supersecretkey")
	viper.SetDefault("JWT_EXPIRATION_HOURS", 24)

	viper.SetDefault("WORKERS_POOL_SIZE", 5)
	viper.SetDefault("WORKERS_BATCH_SIZE", 10)
	viper.SetDefault("WORKERS_BATCH_TIMEOUT", 5*time.Second)
}
