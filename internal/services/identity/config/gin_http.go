package config

type GinHttpConfig struct {
	// Host string `mapstructure:"gin_http_host"`
	Port int `mapstructure:"gin_http_port"`
}
