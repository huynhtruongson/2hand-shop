package connection

import "fmt"

type RabbitMQConnectionConfiguration struct {
	HostName    string `json:"host_name"`
	VirtualHost string `json:"virtual_host"`
	UserName    string `json:"username"`
	Password    string `json:"password"`
}

func (c *RabbitMQConnectionConfiguration) AmqpEndpoint() string {
	return fmt.Sprintf("amqp://%s:%s@%s/%s", c.UserName, c.Password, c.HostName, c.VirtualHost)
}
