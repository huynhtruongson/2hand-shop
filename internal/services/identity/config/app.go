package config

type AppConfig struct {
	ServiceName string `mapstructure:"service_name"`
	Environment string `mapstructure:"environment"`
}
