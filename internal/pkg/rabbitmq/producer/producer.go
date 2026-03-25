package producer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/connection"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/rabbitmq/message"
	"github.com/rabbitmq/amqp091-go"
)

type Producer interface {
	PublishMessage(ctx context.Context, exchange, routingKey string, msg *message.RabbitMQMessage) error
}
type rabbitMQProducer struct {
	connection connection.IConnection
	config     *RabbitMQProducerConfiguration
	//logger
}

func NewRabbitMQProducer(conn connection.IConnection, config *RabbitMQProducerConfiguration) Producer {
	return &rabbitMQProducer{
		connection: conn,
		config:     config,
	}
}

func (r *rabbitMQProducer) PublishMessage(ctx context.Context, exchange, routingKey string, msg *message.RabbitMQMessage) error {
	if r.connection == nil {
		return errors.New("connection is nil")
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

	publishing := msg.ToPublishing(r.config.AppId)

	if err := channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		true,
		false,
		publishing,
	); err != nil {
		return err
	}
	return confirmMessage(ctx, confirms)
}

func confirmMessage(ctx context.Context, confirms <-chan amqp091.Confirmation) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return errors.New("message is not acked")
		} else {
			fmt.Println("message is acked")
			return nil
		}
	case <-ctx.Done():
		return errors.New("confirmation timeout")
	}
}

var _ Producer = (*rabbitMQProducer)(nil)
