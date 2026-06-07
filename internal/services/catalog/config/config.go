package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config aggregates all sub-configuration structs for the catalog service.
type Config struct {
	App           AppConfig           `mapstructure:",squash"`
	Logger        LoggerConfig        `mapstructure:",squash"`
	Keycloak      KeycloakConfig      `mapstructure:",squash"`
	Postgres      PostgresConfig      `mapstructure:",squash"`
	GinHttp       GinHttpConfig       `mapstructure:",squash"`
	RabbitMQ      RabbitMQConfig      `mapstructure:",squash"`
	Elasticsearch ElasticsearchConfig `mapstructure:",squash"`
}

// KeycloakConfig holds the Keycloak settings used by the auth middleware.
type KeycloakConfig struct {
	Realm        string `mapstructure:"keycloak_realm"`
	BaseURL      string `mapstructure:"keycloak_base_url"`
	ClientID     string `mapstructure:"keycloak_client_id"`
	ClientSecret string `mapstructure:"keycloak_client_secret"`
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
