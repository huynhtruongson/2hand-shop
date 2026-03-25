package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:",squash"`
	Logger   LoggerConfig   `mapstructure:",squash"`
	Cognito  CognitoConfig  `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	GinHttp  GinHttpConfig  `mapstructure:",squash"`
}

func Load() (Config, error) {
	v := viper.New()

	v.SetConfigFile("/config/.env") // Docker path (mounted volume)
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
