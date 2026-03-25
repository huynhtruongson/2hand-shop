package manager

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
)

type RabbitMQManagerConfiguration struct {
	ProducerConfiguration  *producer.RabbitMQProducerConfiguration
	ConsumerConfigurations []*consumer.RabbitMQConsumerConfiguration
}

type RabbitMQManagerConfigurationBuilderFunc func(builder RabbitMQManagerConfigurationBuilder)
