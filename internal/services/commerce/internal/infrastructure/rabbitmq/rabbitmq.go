package rabbitmq

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	mqconsumer "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	mqproducer "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/commerce/internal/application"
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
			pb.WithAppId("commerce").
				WithExchangeOptions(&mqproducer.ExchangeOptions{
					Durable:    true,
					AutoDelete: false,
				}).
				WithExchanges(mqproducer.Exchange{
					Name: "commerce.events",
					Type: mqproducer.ExchangeTopic,
				})
		})
		b.AddConsumer("commerce-service.commerce.orders.events", func(cb mqconsumer.RabbitMQConsumerConfigurationBuilder) {
			cb.WithBindingOptions(&mqconsumer.RabbitMQBindingOptions{
				Exchange: "commerce.events",
				Key:      "commerce.order.*",
			})
			cb.WithHandler(d.Handle)
		})
		b.AddConsumer("commerce-service.catalog.product.events", func(cb mqconsumer.RabbitMQConsumerConfigurationBuilder) {
			cb.WithBindingOptions(&mqconsumer.RabbitMQBindingOptions{
				Exchange: "catalog.events",
				Key:      "catalog.product.sold",
			})
			cb.WithHandler(d.Handle)
		})
		b.AddConsumer("commerce-service.identity.user.events", func(cb mqconsumer.RabbitMQConsumerConfigurationBuilder) {
			cb.WithBindingOptions(&mqconsumer.RabbitMQBindingOptions{
				Exchange: "identity.events",
				Key:      "identity.user.registered",
			})
			cb.WithHandler(d.Handle)
		})
	})

	return mgr, nil
}

func BuildEventDispatcher(d *dispatcher.EventDispatcher, handlers application.EventHandlers) {

}
