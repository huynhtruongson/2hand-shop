package config

// PostgresConfig holds the PostgreSQL connection parameters.
type PostgresConfig struct {
	Host     string `mapstructure:"postgres_host"`
	Port     int    `mapstructure:"postgres_port"`
	User     string `mapstructure:"postgres_user"`
	Password string `mapstructure:"postgres_password"`
	DBName   string `mapstructure:"postgres_db_name"`
	SSLMode  string `mapstructure:"postgres_ssl_mode"`
}
