package producer

type ExchangeOptions struct {
	Durable    bool
	AutoDelete bool
	Args       map[string]interface{}
}
type Exchange struct {
	Name string
	Type ExchangeType
}
type RabbitMQProducerConfiguration struct {
	AppId           string
	ExchangeOptions *ExchangeOptions
	// Exchanges       []Exchange
}

type RabbitMQProducerConfigurationBuilderFunc func(builder RabbitMQProducerConfigurationBuilder)

func NewDefaultRabbitMQProducerConfiguration() *RabbitMQProducerConfiguration {
	return &RabbitMQProducerConfiguration{
		ExchangeOptions: &ExchangeOptions{
			Durable:    true,
			AutoDelete: false,
		},
	}
}
