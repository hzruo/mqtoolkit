package rabbitmq

import (
	"context"
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer RabbitMQ生产者实现
type Producer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	connected bool
	config    *types.ConnectionConfig
}

// NewProducer 创建RabbitMQ生产者
func NewProducer() mq.Producer {
	return &Producer{}
}

// Connect 连接到RabbitMQ
func (p *Producer) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeRabbitMQ {
		return utils.NewValidationError("Invalid MQ type for RabbitMQ producer", string(config.Type))
	}

	p.config = config

	// 构建连接URL
	vhost := config.VHost
	if vhost == "" {
		vhost = "/"
	}

	var url string
	if config.Username != "" && config.Password != "" {
		url = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
			config.Username, config.Password, config.Host, config.Port, vhost)
	} else {
		url = fmt.Sprintf("amqp://%s:%d%s", config.Host, config.Port, vhost)
	}

	// 建立连接
	conn, err := amqp.Dial(url)
	if err != nil {
		return utils.NewConnectionError("Failed to connect to RabbitMQ", err)
	}

	// 创建通道
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return utils.NewConnectionError("Failed to create channel", err)
	}

	p.conn = conn
	p.channel = channel
	p.connected = true

	return nil
}

// Produce 发送单条消息
func (p *Producer) Produce(ctx context.Context, req *types.ProduceRequest) error {
	if !p.connected || p.channel == nil {
		return utils.NewConnectionError("Producer not connected", nil)
	}

	if !utils.IsValidTopic(req.Topic) {
		return utils.NewValidationError("Invalid queue name", req.Topic)
	}

	// 声明队列（确保队列存在）
	_, err := p.channel.QueueDeclare(
		req.Topic, // 队列名称
		true,      // 持久化
		false,     // 自动删除
		false,     // 排他性
		false,     // 不等待
		nil,       // 参数
	)
	if err != nil {
		return utils.NewConnectionError("Failed to declare queue", err)
	}

	// 构建消息
	publishing := amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(req.Value),
		DeliveryMode: amqp.Persistent, // 持久化消息
		Timestamp:    time.Now(),
	}

	// 添加Headers
	if req.Headers != nil {
		publishing.Headers = make(amqp.Table)
		for key, value := range req.Headers {
			publishing.Headers[key] = value
		}
	}

	// 发送消息
	err = p.channel.PublishWithContext(
		ctx,
		"",        // exchange
		req.Topic, // routing key (queue name)
		false,     // mandatory
		false,     // immediate
		publishing,
	)

	if err != nil {
		return utils.NewConnectionError("Failed to publish message", err)
	}

	return nil
}

// ProduceBatch 批量发送消息
func (p *Producer) ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) error {
	if !p.connected || p.channel == nil {
		return utils.NewConnectionError("Producer not connected", nil)
	}

	if len(reqs) == 0 {
		return utils.NewValidationError("Empty message batch", "")
	}

	// 逐个发送消息（RabbitMQ没有原生的批量发送API）
	for _, req := range reqs {
		if err := p.Produce(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

// Close 关闭生产者
func (p *Producer) Close() error {
	var lastErr error

	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			lastErr = err
		}
		p.channel = nil
	}

	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			lastErr = err
		}
		p.conn = nil
	}

	p.connected = false
	p.config = nil
	return lastErr
}

// IsConnected 检查连接状态
func (p *Producer) IsConnected() bool {
	if !p.connected || p.conn == nil || p.channel == nil {
		return false
	}

	// 检查连接是否仍然有效
	return !p.conn.IsClosed()
}
