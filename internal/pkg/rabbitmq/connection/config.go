package connection

import "fmt"

type RabbitMQConnectionConfiguration struct {
	Host        string
	Port        int
	VirtualHost string
	User        string
	Password    string
}

func (c *RabbitMQConnectionConfiguration) AmqpEndpoint() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.VirtualHost)
}
