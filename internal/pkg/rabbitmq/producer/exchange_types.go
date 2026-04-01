package producer

import "github.com/rabbitmq/amqp091-go"

type ExchangeType string

const (
	ExchangeFanout ExchangeType = amqp091.ExchangeFanout
	ExchangeDirect ExchangeType = amqp091.ExchangeDirect
	ExchangeTopic  ExchangeType = amqp091.ExchangeTopic
)
