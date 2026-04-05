package rabbitmq

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	mqconsumer "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	mqproducer "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/catalog/internal/application"
)

func NewRabbitMQManager(cfg config.RabbitMQConfig, logger logger.Logger, d *dispatcher.EventDispatcher) (manager.Manager, error) {
	connCfg := &connection.RabbitMQConnectionConfiguration{
		Host:        cfg.Host,
		Port:        cfg.Port,
		VirtualHost: cfg.VirtualHost,
		User:        cfg.User,
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

	mgr := manager.NewRabbitMQManager(logger, consumerConn, producerConn, func(b manager.RabbitMQManagerConfigurationBuilder) {
		b.AddProducer(func(pb mqproducer.RabbitMQProducerConfigurationBuilder) {
			pb.WithAppId("catalog").
				WithExchangeOptions(&mqproducer.ExchangeOptions{
					Durable:    true,
					AutoDelete: false,
				}).
				WithExchanges(mqproducer.Exchange{
					Name: "catalog.events",
					Type: mqproducer.ExchangeTopic,
				})
		})
		b.AddConsumer("catalog-service.catalog.products.events", func(cb mqconsumer.RabbitMQConsumerConfigurationBuilder) {
			cb.WithBindingOptions(&mqconsumer.RabbitMQBindingOptions{
				Exchange: "catalog.events",
				Key:      "catalog.product.*",
			})
			cb.WithHandler(d.Handle)
		})
	})

	return mgr, nil
}

func BuildEventDispatcher(d *dispatcher.EventDispatcher, handlers application.EventHandlers) {
	d.Register("catalog.product.created", dispatcher.NewTypedHandler(handlers.OnProductCreated))
	d.Register("catalog.product.updated", dispatcher.NewTypedHandler(handlers.OnProductUpdated))
	d.Register("catalog.product.deleted", dispatcher.NewTypedHandler(handlers.OnProductDeleted))
}
