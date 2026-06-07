package rabbitmq

import (
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	mqconsumer "github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/dispatcher"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/manager"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/config"
	"github.com/huynhtruongson/2hand-shop/internal/services/identity/internal/application"
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
		b.AddConsumer("identity-service.keycloak.registration.events", func(cb mqconsumer.RabbitMQConsumerConfigurationBuilder) {
			cb.WithBindingOptions(&mqconsumer.RabbitMQBindingOptions{
				Exchange: "keycloak.events",
				Key:      "KK.EVENT.CLIENT.2hand-shop.SUCCESS.2hand-shop-client.REGISTER",
			})
			cb.WithHandler(d.Handle)
		})
	})

	return mgr, nil
}

func BuildEventDispatcher(d *dispatcher.EventDispatcher, handlers application.EventHandlers) {
	d.Register("KK.EVENT.CLIENT.2hand-shop.SUCCESS.2hand-shop-client.REGISTER", newTypedHandler(handlers.OnKeycloakUserRegistration))
}
