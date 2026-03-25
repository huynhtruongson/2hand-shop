package consumer

type RabbitMQConsumerConfigurationBuilder interface {
	WithName(name string) RabbitMQConsumerConfigurationBuilder
	WithConsumerId(id string) RabbitMQConsumerConfigurationBuilder
	WithHandler(handler RabbitMQConsumerHandler) RabbitMQConsumerConfigurationBuilder
	WithConcurrencyLimit(limit int) RabbitMQConsumerConfigurationBuilder
	WithPrefetchCount(count int) RabbitMQConsumerConfigurationBuilder
	WithAutoAck(autoAck bool) RabbitMQConsumerConfigurationBuilder
	WithNoLocal(noLocal bool) RabbitMQConsumerConfigurationBuilder
	WithNoWait(noWait bool) RabbitMQConsumerConfigurationBuilder
	WithQueueOptions(options *RabbitMQQueueOptions) RabbitMQConsumerConfigurationBuilder
	WithBindingOptions(options *RabbitMQBindingOptions) RabbitMQConsumerConfigurationBuilder
	Build() *RabbitMQConsumerConfiguration
}

type rabbitMQConsumerConfigurationBuilder struct {
	config *RabbitMQConsumerConfiguration
}

func NewRabbitMQConsumerConfigurationBuilder(name string) RabbitMQConsumerConfigurationBuilder {
	return &rabbitMQConsumerConfigurationBuilder{
		config: NewDefaultRabbitMQConsumerConfiguration(name),
	}
}

func (r *rabbitMQConsumerConfigurationBuilder) WithAutoAck(autoAck bool) RabbitMQConsumerConfigurationBuilder {
	r.config.AutoAck = autoAck
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithBindingOptions(options *RabbitMQBindingOptions) RabbitMQConsumerConfigurationBuilder {
	r.config.BindingOptions = options
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithConcurrencyLimit(limit int) RabbitMQConsumerConfigurationBuilder {
	r.config.ConcurrencyLimit = limit
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithConsumerId(id string) RabbitMQConsumerConfigurationBuilder {
	r.config.ConsumerId = id
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithHandler(handler RabbitMQConsumerHandler) RabbitMQConsumerConfigurationBuilder {
	r.config.Handler = handler
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithName(name string) RabbitMQConsumerConfigurationBuilder {
	r.config.Name = name
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithNoLocal(noLocal bool) RabbitMQConsumerConfigurationBuilder {
	r.config.NoLocal = noLocal
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithNoWait(noWait bool) RabbitMQConsumerConfigurationBuilder {
	r.config.NoWait = noWait
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithPrefetchCount(count int) RabbitMQConsumerConfigurationBuilder {
	r.config.PrefetchCount = count
	return r
}

func (r *rabbitMQConsumerConfigurationBuilder) WithQueueOptions(options *RabbitMQQueueOptions) RabbitMQConsumerConfigurationBuilder {
	r.config.QueueOptions = options
	return r
}
func (r *rabbitMQConsumerConfigurationBuilder) Build() *RabbitMQConsumerConfiguration {
	return r.config
}

var _ RabbitMQConsumerConfigurationBuilder = (*rabbitMQConsumerConfigurationBuilder)(nil)
