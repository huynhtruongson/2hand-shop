package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/consumer"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/producer"
)

type Manager interface {
	Producer() producer.Producer
	Start(ctx context.Context) error
	Stop() error
}

type rabbitMQManager struct {
	consumerConnection connection.IConnection
	producerConnection connection.IConnection
	consumers          []consumer.Consumer
	producer           producer.Producer
	config             *RabbitMQManagerConfiguration
	logger             logger.Logger
}

func NewRabbitMQManager(logger logger.Logger, consumerConn, produercerConn connection.IConnection, builderFn RabbitMQManagerConfigurationBuilderFunc) Manager {
	builder := NewRabbitMQManagerConfigurationBuilder()
	if builderFn != nil {
		builderFn(builder)
	}
	config := builder.Build()

	manager := &rabbitMQManager{
		logger:             logger,
		consumerConnection: consumerConn,
		producerConnection: produercerConn,
		config:             config,
	}

	for _, consumerConfig := range manager.config.ConsumerConfigurations {
		consumer := consumer.NewRabbitMQConsumer(manager.logger, manager.consumerConnection, consumerConfig)
		manager.consumers = append(manager.consumers, consumer)
	}
	if manager.config.ProducerConfiguration != nil {
		manager.producer = producer.NewRabbitMQProducer(manager.logger, manager.producerConnection, manager.config.ProducerConfiguration)
	}
	return manager
}

func (r *rabbitMQManager) Producer() producer.Producer {
	return r.producer
}

func (r *rabbitMQManager) Start(ctx context.Context) error {
	if r.producer != nil {
		if err := r.producer.ExchangesDeclare(ctx); err != nil {
			return err
		}
	}

	for _, consumer := range r.consumers {
		if err := consumer.Start(ctx); err != nil {
			r.logger.Error(fmt.Sprintf("consumer %s started failed", consumer.GetName()), "error", err)
			err2 := consumer.Stop()
			if err2 != nil {
				return err2
			}
			return err
		}
	}
	r.logger.Info("rabbitmq is running")
	return nil

}

func (r *rabbitMQManager) Stop() error {
	wg := sync.WaitGroup{}
	for _, consumer := range r.consumers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := consumer.Stop(); err != nil {
				r.logger.Error("error to stop consuming", "error", err)
			}
		}()
	}
	wg.Wait()
	// if r.consumerConnection != nil {
	// 	r.consumerConnection.Close()
	// }
	// if r.producerConnection != nil {
	// 	r.producerConnection.Close()
	// }
	return nil
}
