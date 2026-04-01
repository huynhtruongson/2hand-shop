package producer

type RabbitMQProducerConfigurationBuilder interface {
	WithAppId(appId string) RabbitMQProducerConfigurationBuilder
	WithExchangeOptions(options *ExchangeOptions) RabbitMQProducerConfigurationBuilder
	WithExchanges(exchanges ...Exchange) RabbitMQProducerConfigurationBuilder
	Build() *RabbitMQProducerConfiguration
}
type rabbitMQProducerConfigurationBuilder struct {
	config *RabbitMQProducerConfiguration
}

func NewRabbitMQProducerConfigurationBuilder() RabbitMQProducerConfigurationBuilder {
	return &rabbitMQProducerConfigurationBuilder{
		config: NewDefaultRabbitMQProducerConfiguration(),
	}
}

func (r *rabbitMQProducerConfigurationBuilder) Build() *RabbitMQProducerConfiguration {
	return r.config
}

func (r *rabbitMQProducerConfigurationBuilder) WithAppId(appId string) RabbitMQProducerConfigurationBuilder {
	r.config.AppId = appId
	return r
}

func (r *rabbitMQProducerConfigurationBuilder) WithExchangeOptions(options *ExchangeOptions) RabbitMQProducerConfigurationBuilder {
	r.config.ExchangeOptions = options
	return r
}

func (r *rabbitMQProducerConfigurationBuilder) WithExchanges(exchanges ...Exchange) RabbitMQProducerConfigurationBuilder {
	r.config.Exchanges = exchanges
	return r
}

var _ RabbitMQProducerConfigurationBuilder = (*rabbitMQProducerConfigurationBuilder)(nil)
