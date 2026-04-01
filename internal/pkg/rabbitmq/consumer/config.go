package consumer

import (
	"context"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
)

type RabbitMQConsumerHandler func(ctx context.Context, msg *types.DeliveryMessage) error

type RabbitMQConsumerConfigurationBuilderFunc func(builder RabbitMQConsumerConfigurationBuilder)

type RabbitMQQueueOptions struct {
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	Args       map[string]any
}

type RabbitMQBindingOptions struct {
	Exchange string
	Key      string
	Args     map[string]interface{}
}

type RabbitMQConsumerConfiguration struct {
	Name             string
	ConsumerId       string
	Handler          RabbitMQConsumerHandler
	ConcurrencyLimit int
	PrefetchCount    int
	AutoAck          bool
	NoLocal          bool
	NoWait           bool
	QueueOptions     *RabbitMQQueueOptions
	BindingOptions   *RabbitMQBindingOptions
}

func NewDefaultRabbitMQConsumerConfiguration(name string) *RabbitMQConsumerConfiguration {
	return &RabbitMQConsumerConfiguration{
		Name:             name,
		ConsumerId:       "",
		ConcurrencyLimit: 1,
		PrefetchCount:    4,
		NoLocal:          false,
		NoWait:           false,
		AutoAck:          false,
		QueueOptions: &RabbitMQQueueOptions{
			Durable:    true,
			Exclusive:  false,
			AutoDelete: false,
		},
	}
}
