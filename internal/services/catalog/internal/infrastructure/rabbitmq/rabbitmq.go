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

func NewRabbitMQManager(cfg config.RabbitMQConfig, logger logger.Logger, app application.Application) (manager.Manager, error) {
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

	// Build the dispatcher using the event handlers injected from the application layer.
	dispatcher := buildEventDispatcher(logger, app.EventHandlers)

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
			cb.WithHandler(dispatcher.Handle)
		})
	})

	return mgr, nil
}

func buildEventDispatcher(log logger.Logger, handlers application.EventHandlers) *dispatcher.EventDispatcher {
	b := dispatcher.NewBuilder(log)
	dispatcher.Register(b, "catalog.product.created", handlers.OnProductCreated)

	d, err := b.Build()
	if err != nil {
		panic("rabbitmq: failed to build event dispatcher: " + err.Error())
	}
	return d
}
