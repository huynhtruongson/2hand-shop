package consumer

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/avast/retry-go"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"

	"github.com/rabbitmq/amqp091-go"
)

const (
	retryAttempts = 3
	retryDelay    = 300 * time.Millisecond
)

var retryOptions = []retry.Option{
	retry.Attempts(retryAttempts),
	retry.Delay(retryDelay),
	retry.DelayType(retry.BackOffDelay),
}

type Consumer interface {
	Start(ctx context.Context) error
	Stop() error
	GetName() string
}
type rabbitMQConsumer struct {
	config     *RabbitMQConsumerConfiguration
	connection connection.IConnection
	channel    *amqp091.Channel
	handler    RabbitMQConsumerHandler
	logger     logger.Logger
}

var _ Consumer = (*rabbitMQConsumer)(nil)

func NewRabbitMQConsumer(conn connection.IConnection, config *RabbitMQConsumerConfiguration) Consumer {
	return &rabbitMQConsumer{
		config:     config,
		connection: conn,
		handler:    config.Handler,
	}
}

// GetName implements Consumer.
func (r *rabbitMQConsumer) GetName() string {
	return r.config.Name
}

// Start implements Consumer.
func (r *rabbitMQConsumer) Start(ctx context.Context) error {
	switch {
	case r.connection == nil:
		return errors.New("connection is nil")
	case r.config.QueueOptions == nil:
		return errors.New("queue options is nil")
	case r.config.BindingOptions == nil:
		return errors.New("binding options is nil")
	case r.config.Handler == nil:
		return errors.New("handle is nil")
	}

	r.reConsumeOnDropConnection(ctx)

	ch, err := r.connection.Channel()
	if err != nil {
		return err
	}
	r.channel = ch

	// The prefetch count tells the Rabbit connection how many messages to retrieve from the server per request.
	prefetchCount := r.config.ConcurrencyLimit * r.config.PrefetchCount
	if err := r.channel.Qos(prefetchCount, 0, false); err != nil {
		return err
	}

	_, err = r.channel.QueueDeclare(
		r.config.Name,
		r.config.QueueOptions.Durable,
		r.config.QueueOptions.AutoDelete,
		r.config.QueueOptions.Exclusive,
		r.config.NoWait,
		r.config.QueueOptions.Args,
	)
	if err != nil {
		return err
	}
	if err := r.channel.QueueBind(
		r.config.Name,
		r.config.BindingOptions.Key,
		r.config.BindingOptions.Exchange,
		r.config.NoWait,
		r.config.BindingOptions.Args,
	); err != nil {
		return err
	}
	msgs, err := r.channel.Consume(
		r.config.Name,
		r.config.ConsumerId,
		r.config.AutoAck,
		r.config.QueueOptions.Exclusive,
		r.config.NoLocal,
		r.config.NoWait,
		nil,
	)
	if err != nil {
		return err
	}
	chClosed := make(chan *amqp091.Error, 1)
	r.channel.NotifyClose(chClosed)
	for i := 0; i < r.config.ConcurrencyLimit; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					r.logger.Info("Shutting down consumer")
					return
				case err := <-chClosed:
					r.logger.Info("Channel closed with error", "error", err)
					return
					// chClosed = make(chan *amqp091.Error, 1)
					// ch.NotifyClose(chClosed)
				case msg, ok := <-msgs:
					if !ok {
						r.logger.Info("Consumer connection dropped")
						return
					}
					var message = msg
					r.handleReceiveMessage(ctx, &message)
				}
			}
		}()
	}
	return nil
}

func (r *rabbitMQConsumer) Stop() error {
	if r.channel != nil && r.channel.IsClosed() == false {
		if err := r.channel.Cancel(r.config.ConsumerId, false); err != nil {
			return err
		}
		return r.channel.Close()
	}
	return nil
}

func (r *rabbitMQConsumer) reConsumeOnDropConnection(ctx context.Context) {
	go func() {
		for {
			select {
			case reconnect := <-r.connection.ReconnectedChannel():
				if reflect.ValueOf(reconnect).IsValid() {
					r.logger.Info("reconsume_on_drop_connection started")
					if err := r.Start(ctx); err != nil {
						r.logger.Error("reconsume_on_drop_connection finished with error", "error", err)
						continue
					}
					r.logger.Info("reconsume_on_drop_connection finished successfully")
					return
				}
			}
		}
	}()
}

func (r *rabbitMQConsumer) handleReceiveMessage(ctx context.Context, msg *amqp091.Delivery) {
	err := retry.Do(func() error {
		return r.handler(ctx, msg)
	}, append(retryOptions, retry.Context(ctx))...)
	if err != nil {
		r.logger.Error("[rabbitMQConsumer.Handle] error in handling consume message of RabbitmqMQ, prepare for nacking message")
		if r.config.AutoAck == false {
			if err := msg.Nack(false, true); err != nil {
				r.logger.Error("error in sending Nack to RabbitMQ consumer", "error", err)
				return
			}
		}
	} else {
		if r.config.AutoAck == false {
			if err := msg.Ack(false); err != nil {
				r.logger.Error("error in sending Ack to RabbitMQ consumer", "error", err)
				return
			}
		}
	}
}
