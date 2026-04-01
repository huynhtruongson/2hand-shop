package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config aggregates all sub-configuration structs for the catalog service.
type Config struct {
	App      AppConfig      `mapstructure:",squash"`
	Logger   LoggerConfig   `mapstructure:",squash"`
	Cognito  CognitoConfig  `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	GinHttp  GinHttpConfig  `mapstructure:",squash"`
	RabbitMQ RabbitMQConfig `mapstructure:",squash"`
}

// CognitoConfig holds the AWS Cognito user pool settings used by the auth middleware.
type CognitoConfig struct {
	Region     string `mapstructure:"cognito_region"`
	UserPoolID string `mapstructure:"cognito_user_pool_id"`
	ClientID   string `mapstructure:"cognito_client_id"`
}

// Load reads configuration from /config/.env using Viper.
func Load() (Config, error) {
	v := viper.New()

	v.SetConfigFile("/config/.env")
	v.SetConfigType("env")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unable to decode config: %w", err)
	}

	return cfg, nil
}
