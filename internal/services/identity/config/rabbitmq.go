package config

// RabbitMQConfig holds the RabbitMQ connection parameters.
type RabbitMQConfig struct {
	Host        string `mapstructure:"rabbitmq_host"`
	Port        int    `mapstructure:"rabbitmq_port"`
	VirtualHost string `mapstructure:"rabbitmq_vhost"`
	User        string `mapstructure:"rabbitmq_user"`
	Password    string `mapstructure:"rabbitmq_password"`
}
