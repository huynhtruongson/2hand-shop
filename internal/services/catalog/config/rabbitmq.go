package config

// RabbitMQConfig holds the RabbitMQ connection parameters.
type RabbitMQConfig struct {
	HostName    string `mapstructure:"rabbitmq_host_name"`
	VirtualHost string `mapstructure:"rabbitmq_virtual_host"`
	UserName    string `mapstructure:"rabbitmq_user_name"`
	Password    string `mapstructure:"rabbitmq_password"`
}
