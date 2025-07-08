package rabbitmq

import (
	"context"
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer RabbitMQ消费者实现
type Consumer struct {
	conn             *amqp.Connection
	channel          *amqp.Channel
	connected        bool
	config           *types.ConnectionConfig
	subscribedQueues []string // 保存订阅的队列
}

// NewConsumer 创建RabbitMQ消费者
func NewConsumer() mq.Consumer {
	return &Consumer{}
}

// Connect 连接到RabbitMQ
func (c *Consumer) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeRabbitMQ {
		return utils.NewValidationError("Invalid MQ type for RabbitMQ consumer", string(config.Type))
	}

	c.config = config
	vhost := config.VHost
	if vhost == "" {
		vhost = "/"
	}

	url := fmt.Sprintf("amqp://%s:%s@%s:%d%s", config.Username, config.Password, config.Host, config.Port, vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return utils.NewConnectionError("Failed to connect to RabbitMQ", err)
	}
	c.conn = conn
	c.connected = true
	return nil
}

// Subscribe 订阅队列
func (c *Consumer) Subscribe(ctx context.Context, req *types.ConsumeRequest) error {
	if !c.connected || c.conn == nil {
		return utils.NewConnectionError("Consumer not connected", nil)
	}

	if len(req.Topics) == 0 {
		return utils.NewValidationError("No queues specified for subscription", "")
	}

	// 如果已有通道，先关闭
	if c.channel != nil {
		c.channel.Close()
	}

	channel, err := c.conn.Channel()
	if err != nil {
		return utils.NewConnectionError("Failed to create channel", err)
	}
	c.channel = channel

	// 声明队列并保存
	c.subscribedQueues = []string{}
	for _, queueName := range req.Topics {
		_, err := c.channel.QueueDeclare(
			queueName, // name
			true,      // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			return utils.NewConnectionError(fmt.Sprintf("Failed to declare queue %s", queueName), err)
		}
		c.subscribedQueues = append(c.subscribedQueues, queueName)
	}
	return nil
}

// Consume 消费消息
func (c *Consumer) Consume(ctx context.Context, handler mq.MessageHandler) error {
	if !c.connected || c.channel == nil {
		return utils.NewConnectionError("Consumer not subscribed. Call Subscribe first.", nil)
	}

	if len(c.subscribedQueues) == 0 {
		return utils.NewValidationError("No queues subscribed", "")
	}

	if err := c.channel.Qos(1, 0, false); err != nil {
		return utils.NewConnectionError("Failed to set QoS", err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, queueName := range c.subscribedQueues {
		wg.Add(1)
		go func(qName string) {
			defer wg.Done()
			
			msgs, err := c.channel.Consume(
				qName, "", false, false, false, false, nil,
			)
			if err != nil {
				fmt.Printf("Failed to start consuming from queue %s: %v\n", qName, err)
				cancel() // 取消其他 goroutines
				return
			}

			for {
				select {
				case <-ctx.Done():
					return
				case delivery, ok := <-msgs:
					if !ok {
						return
					}
					
					msg := &types.Message{
						ID:        utils.GenerateID(),
						Topic:     qName,
						Key:       delivery.RoutingKey,
						Value:     string(delivery.Body),
						Timestamp: delivery.Timestamp,
					}
					if delivery.Headers != nil {
						msg.Headers = make(map[string]string)
						for key, value := range delivery.Headers {
							msg.Headers[key] = fmt.Sprintf("%v", value)
						}
					}

					if err := handler(msg); err != nil {
						fmt.Printf("Failed to handle message from %s: %v. Nacking.\n", qName, err)
						delivery.Nack(false, true)
					} else {
						delivery.Ack(false)
					}
				}
			}
		}(queueName)
	}

	wg.Wait()
	return ctx.Err()
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.connected = false
	var lastErr error
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			lastErr = err
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			lastErr = err
		}
	}
	c.subscribedQueues = nil
	return lastErr
}

// IsConnected 检查连接状态
func (c *Consumer) IsConnected() bool {
	return c.connected && c.conn != nil && !c.conn.IsClosed()
}

