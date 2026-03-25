package connection

import (
	"fmt"

	"github.com/huynhtruongson/2hand-shop/internal/pkg/logger"
	"github.com/huynhtruongson/2hand-shop/internal/pkg/utils"
	"github.com/rabbitmq/amqp091-go"
)

type rabbitMQConnection struct {
	*amqp091.Connection
	logger            logger.Logger
	cfg               *RabbitMQConnectionConfiguration
	isConnected       bool
	errConnectionChan chan error
	reconnectedChan   chan struct{}
}

type IConnection interface {
	IsClosed() bool
	IsConnected() bool
	Channel() (*amqp091.Channel, error)
	Close() error
	ReConnect() error
	NotifyClose(receiver chan *amqp091.Error) chan *amqp091.Error
	Raw() *amqp091.Connection
	ErrorConnectionChannel() chan error
	ReconnectedChannel() chan struct{}
}

var _ IConnection = (*rabbitMQConnection)(nil)

func NewRabbitMQConnection(config *RabbitMQConnectionConfiguration, logger logger.Logger) (*rabbitMQConnection, error) {
	c := &rabbitMQConnection{
		cfg:               config,
		logger:            logger,
		errConnectionChan: make(chan error),
		reconnectedChan:   make(chan struct{}),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	go c.handleReconnecting()
	return c, nil
}

func (c *rabbitMQConnection) ErrorConnectionChannel() chan error {
	return c.errConnectionChan
}

func (c *rabbitMQConnection) IsConnected() bool {
	return c.isConnected
}

// Raw implements IConnection.
func (c *rabbitMQConnection) Raw() *amqp091.Connection {
	return c.Connection
}

func (c *rabbitMQConnection) ReConnect() error {
	if c.Connection.IsClosed() == false {
		return nil
	}
	return c.connect()
}

func (c *rabbitMQConnection) ReconnectedChannel() chan struct{} {
	return c.reconnectedChan
}

func (c *rabbitMQConnection) connect() error {
	conn, err := amqp091.Dial(c.cfg.AmqpEndpoint())
	if err != nil {
		return fmt.Errorf("Error in connecting to rabbitmq with host: %s, err: %w", c.cfg.AmqpEndpoint(), err)
	}
	c.Connection = conn
	c.isConnected = true
	notifyClose := c.Connection.NotifyClose(make(chan *amqp091.Error))

	go func() {
		defer utils.HandlePanic()
		chanErr := <-notifyClose
		c.errConnectionChan <- chanErr
		c.isConnected = false
	}()
	return nil
}

func (c *rabbitMQConnection) handleReconnecting() {
	defer utils.HandlePanic()
	for {
		select {
		case err := <-c.errConnectionChan:
			if err != nil {
				c.logger.Info(("Rabbitmq Connection Reconnecting started"))
				err := c.connect()
				if err != nil {
					c.logger.Error("Rabbitmq Connection reconnected error", "error", err)
					continue
				}
				c.logger.Info("Rabbitmq Connection Reconnected")
				c.reconnectedChan <- struct{}{}
				c.isConnected = true
			}
		}
	}
}
