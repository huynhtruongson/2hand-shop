package rabbitmq

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
)

// NewRabbitMQManager creates and configures the RabbitMQ manager for the catalog service.
// It establishes two connections (one for consumers, one for the producer) and wires
// the producer with the catalog.events exchange. No consumers are registered here;
// add them via manager.AddConsumer in a follow-up step.
func NewRabbitMQManager(cfg config.RabbitMQConfig, logger logger.Logger) (manager.Manager, error) {
	connCfg := &connection.RabbitMQConnectionConfiguration{
		HostName:    cfg.HostName,
		VirtualHost: cfg.VirtualHost,
		UserName:    cfg.UserName,
		Password:    cfg.Password,
	}

	consumerConn, err := connection.NewRabbitMQConnection(connCfg, logger)
	if err != nil {
		return nil, err
	}

	producerConn, err := connection.NewRabbitMQConnection(connCfg, logger)
	if err != nil {
		return nil, err
	}

	mgr := manager.NewRabbitMQManager(consumerConn, producerConn, func(b manager.RabbitMQManagerConfigurationBuilder) {
		b.AddProducer(func(pb producer.RabbitMQProducerConfigurationBuilder) {
			pb.WithAppId("catalog").
				WithExchangeOptions(&producer.ExchangeOptions{
					Durable:    true,
					AutoDelete: false,
				}).
				WithExchanges(producer.Exchange{
					Name: "catalog.events",
					Type: producer.ExchangeTopic,
				})
		})
	})

	return mgr, nil
}
