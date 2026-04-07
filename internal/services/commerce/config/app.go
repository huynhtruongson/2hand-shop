package config

// AppConfig holds service-level metadata.
type AppConfig struct {
	ServiceName string `mapstructure:"service_name"`
	Environment string `mapstructure:"environment"`
}
