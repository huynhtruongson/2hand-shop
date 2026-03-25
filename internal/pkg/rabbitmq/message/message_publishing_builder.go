package message

import (
	"time"

	"github.com/rabbitmq/amqp091-go"
	uuid "github.com/satori/go.uuid"
)

type RabbitMQPublishingBuilder interface {
	WithCorrelationId(id string) RabbitMQPublishingBuilder
	WithHeaders(table map[string]interface{}) RabbitMQPublishingBuilder
	WithType(name string) RabbitMQPublishingBuilder
	WithDeliveryMode(mode uint8) RabbitMQPublishingBuilder
	WithReplyTo(replyTo string) RabbitMQPublishingBuilder
	WithPriority(priority uint8) RabbitMQPublishingBuilder
	WithContentEncoding(encoding string) RabbitMQPublishingBuilder
	WithExpiration(expiration string) RabbitMQPublishingBuilder
	Build() *amqp091.Publishing
}

type rabbitMQPublishingBuilder struct {
	publishing *amqp091.Publishing
}

func NewRabbitMQPublishingBuilder() RabbitMQPublishingBuilder {
	publishing := &amqp091.Publishing{
		MessageId:    uuid.NewV4().String(),
		Timestamp:    time.Now().UTC(),
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
	}
	return &rabbitMQPublishingBuilder{publishing: publishing}
}

func (r *rabbitMQPublishingBuilder) Build() *amqp091.Publishing {
	return r.publishing
}

func (r *rabbitMQPublishingBuilder) WithContentEncoding(encoding string) RabbitMQPublishingBuilder {
	r.publishing.ContentEncoding = encoding
	return r
}

func (r *rabbitMQPublishingBuilder) WithCorrelationId(id string) RabbitMQPublishingBuilder {
	r.publishing.CorrelationId = id
	return r
}

func (r *rabbitMQPublishingBuilder) WithDeliveryMode(mode uint8) RabbitMQPublishingBuilder {
	r.publishing.DeliveryMode = mode
	return r
}

func (r *rabbitMQPublishingBuilder) WithExpiration(expiration string) RabbitMQPublishingBuilder {
	r.publishing.Expiration = expiration
	return r
}

func (r *rabbitMQPublishingBuilder) WithHeaders(table map[string]interface{}) RabbitMQPublishingBuilder {
	r.publishing.Headers = amqp091.Table(table)
	return r
}

func (r *rabbitMQPublishingBuilder) WithPriority(priority uint8) RabbitMQPublishingBuilder {
	r.publishing.Priority = priority
	return r
}

func (r *rabbitMQPublishingBuilder) WithReplyTo(replyTo string) RabbitMQPublishingBuilder {
	r.publishing.ReplyTo = replyTo
	return r
}

func (r *rabbitMQPublishingBuilder) WithType(typeName string) RabbitMQPublishingBuilder {
	r.publishing.Type = typeName
	return r
}

var _ RabbitMQPublishingBuilder = (*rabbitMQPublishingBuilder)(nil)
