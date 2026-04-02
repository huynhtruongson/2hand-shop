package producer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/types"
	"github.com/rabbitmq/amqp091-go"
)

type Producer interface {
	PublishMessage(ctx context.Context, message types.DomainEvent, opts ...types.MessageOption) error
	ExchangesDeclare(ctx context.Context) error
}
type rabbitMQProducer struct {
	connection connection.IConnection
	config     *RabbitMQProducerConfiguration
	logger     logger.Logger
}

func NewRabbitMQProducer(conn connection.IConnection, logger logger.Logger, config *RabbitMQProducerConfiguration) Producer {
	return &rabbitMQProducer{
		connection: conn,
		config:     config,
		logger:     logger,
	}
}

func (r *rabbitMQProducer) PublishMessage(ctx context.Context, msg types.DomainEvent, opts ...types.MessageOption) error {
	if r.connection == nil {
		return errors.New("connection is nil")
	}

	envelope := types.NewEventEnvelope(msg.EventType(), msg, msg.CorrelationID())

	rabbitmqMsg, err := types.NewRabbitMQMessage(envelope, opts...)
	if err != nil {
		return fmt.Errorf("build RabbitMQ message: %w", err)
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	if err := channel.Confirm(false); err != nil {
		return err
	}
	confirms := make(chan amqp091.Confirmation, 1)
	channel.NotifyPublish(confirms)

	publishing := rabbitmqMsg.ToPublishing(r.config.AppId)

	if err := channel.PublishWithContext(
		ctx,
		msg.Exchange(),
		msg.EventType(),
		true,
		false,
		publishing,
	); err != nil {
		return err
	}
	return r.confirmMessage(ctx, confirms)
}

func (r *rabbitMQProducer) confirmMessage(ctx context.Context, confirms <-chan amqp091.Confirmation) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return errors.New("message is not acked")
		} else {
			r.logger.Info("message is acked")
			return nil
		}
	case <-ctx.Done():
		return errors.New("confirmation timeout")
	}
}

func (r *rabbitMQProducer) ExchangesDeclare(ctx context.Context) error {
	if r.connection == nil {
		return errors.New("connection is nil")
	}
	if len(r.config.Exchanges) == 0 {
		return nil
	}
	ch, err := r.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	for _, ex := range r.config.Exchanges {
		if err := ch.ExchangeDeclare(
			ex.Name,
			string(ex.Type),
			r.config.ExchangeOptions.Durable,
			r.config.ExchangeOptions.AutoDelete,
			false,
			false,
			r.config.ExchangeOptions.Args,
		); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
		r.logger.Info("exchange declared", "exchange", ex.Name, "type", ex.Type)
	}
	return nil
}

var _ Producer = (*rabbitMQProducer)(nil)
