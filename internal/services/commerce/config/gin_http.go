package config

// GinHttpConfig holds the Gin HTTP server settings.
type GinHttpConfig struct {
	Port int `mapstructure:"gin_http_port"`
}
