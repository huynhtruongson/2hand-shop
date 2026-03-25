package manager

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
)

type RabbitMQManagerConfigurationBuilder interface {
	AddConsumer(name string, builderFn consumer.RabbitMQConsumerConfigurationBuilderFunc) RabbitMQManagerConfigurationBuilder
	AddProducer(builderFn producer.RabbitMQProducerConfigurationBuilderFunc) RabbitMQManagerConfigurationBuilder
	Build() *RabbitMQManagerConfiguration
}

type rabbitMQManagerConfigurationBuilder struct {
	consumerBuilders []consumer.RabbitMQConsumerConfigurationBuilder
	producerBuilder  producer.RabbitMQProducerConfigurationBuilder
	config           *RabbitMQManagerConfiguration
}

func NewRabbitMQManagerConfigurationBuilder() RabbitMQManagerConfigurationBuilder {
	return &rabbitMQManagerConfigurationBuilder{
		config: &RabbitMQManagerConfiguration{},
	}
}

func (r *rabbitMQManagerConfigurationBuilder) AddConsumer(name string, builderFn consumer.RabbitMQConsumerConfigurationBuilderFunc) RabbitMQManagerConfigurationBuilder {
	builder := consumer.NewRabbitMQConsumerConfigurationBuilder(name)
	if builderFn != nil {
		builderFn(builder)
	}
	r.consumerBuilders = append(r.consumerBuilders, builder)

	return r
}

func (r *rabbitMQManagerConfigurationBuilder) AddProducer(builderFn producer.RabbitMQProducerConfigurationBuilderFunc) RabbitMQManagerConfigurationBuilder {
	builder := producer.NewRabbitMQProducerConfigurationBuilder()
	if builderFn != nil {
		builderFn(builder)
	}
	r.producerBuilder = builder

	return r
}

func (r *rabbitMQManagerConfigurationBuilder) Build() *RabbitMQManagerConfiguration {
	consumerConfigs := make([]*consumer.RabbitMQConsumerConfiguration, 0, len(r.consumerBuilders))
	for _, builder := range r.consumerBuilders {
		consumerConfigs = append(consumerConfigs, builder.Build())
	}

	var producerConfig *producer.RabbitMQProducerConfiguration
	if r.producerBuilder != nil {
		producerConfig = r.producerBuilder.Build()
	}

	r.config.ConsumerConfigurations = consumerConfigs
	r.config.ProducerConfiguration = producerConfig

	return r.config
}
