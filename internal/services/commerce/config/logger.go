package config

// LoggerConfig controls the structured logger output.
type LoggerConfig struct {
	Level string `mapstructure:"logger_level"`
}
