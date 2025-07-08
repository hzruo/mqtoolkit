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

// Admin RabbitMQ管理客户端实现
type Admin struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	connected bool
	config    *types.ConnectionConfig
}

// NewAdmin 创建RabbitMQ管理客户端
func NewAdmin() mq.Admin {
	return &Admin{}
}

// Connect 连接到RabbitMQ
func (a *Admin) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeRabbitMQ {
		return utils.NewValidationError("Invalid MQ type for RabbitMQ admin", string(config.Type))
	}

	a.config = config

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

	a.conn = conn
	a.channel = channel
	a.connected = true

	return nil
}

// TestConnection 测试连接
func (a *Admin) TestConnection(ctx context.Context) *types.TestResult {
	start := time.Now()

	if !a.connected || a.conn == nil || a.channel == nil {
		return &types.TestResult{
			Success: false,
			Message: "Not connected to RabbitMQ",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 检查连接状态
	if a.conn.IsClosed() {
		return &types.TestResult{
			Success: false,
			Message: "Connection is closed",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 尝试声明一个临时队列来测试连接
	testQueue := fmt.Sprintf("test-queue-%d", time.Now().UnixNano())
	queue, err := a.channel.QueueDeclare(
		testQueue, // 队列名称
		false,     // 不持久化
		true,      // 自动删除
		true,      // 排他性
		false,     // 不等待
		nil,       // 参数
	)
	if err != nil {
		return &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Failed to declare test queue: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 删除测试队列
	_, err = a.channel.QueueDelete(queue.Name, false, false, false)
	if err != nil {
		return &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Failed to delete test queue: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &types.TestResult{
		Success: true,
		Message: "Connected successfully to RabbitMQ",
		Latency: time.Since(start).Milliseconds(),
	}
}

// ListTopics 列出所有队列（RabbitMQ中的"主题"概念对应队列）
func (a *Admin) ListTopics(ctx context.Context) ([]types.TopicInfo, error) {
	if !a.connected || a.channel == nil {
		return nil, utils.NewConnectionError("Not connected to RabbitMQ", nil)
	}

	// RabbitMQ的AMQP协议没有直接的API来列出所有队列
	// 这需要使用RabbitMQ的HTTP管理API
	// 这里返回空列表，实际实现需要使用HTTP API
	return []types.TopicInfo{}, nil
}

// CreateTopic 创建队列
func (a *Admin) CreateTopic(ctx context.Context, topic string, partitions int32, replicas int16) error {
	if !a.connected || a.channel == nil {
		return utils.NewConnectionError("Not connected to RabbitMQ", nil)
	}

	if !utils.IsValidTopic(topic) {
		return utils.NewValidationError("Invalid queue name", topic)
	}

	// 在RabbitMQ中，partitions和replicas参数不适用
	// 这里只创建一个持久化队列
	_, err := a.channel.QueueDeclare(
		topic, // 队列名称
		true,  // 持久化
		false, // 自动删除
		false, // 排他性
		false, // 不等待
		nil,   // 参数
	)

	if err != nil {
		return utils.NewConnectionError("Failed to create queue", err)
	}

	return nil
}

// DeleteTopic 删除队列
func (a *Admin) DeleteTopic(ctx context.Context, topic string) error {
	if !a.connected || a.channel == nil {
		return utils.NewConnectionError("Not connected to RabbitMQ", nil)
	}

	if !utils.IsValidTopic(topic) {
		return utils.NewValidationError("Invalid queue name", topic)
	}

	// 删除队列
	_, err := a.channel.QueueDelete(
		topic, // 队列名称
		false, // 如果未使用
		false, // 如果为空
		false, // 不等待
	)

	if err != nil {
		return utils.NewConnectionError("Failed to delete queue", err)
	}

	return nil
}

// ListConsumerGroups 列出消费组（RabbitMQ没有消费组概念）
func (a *Admin) ListConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error) {
	// RabbitMQ没有消费组的概念，返回空列表
	return []types.ConsumerGroup{}, nil
}

// Close 关闭连接
func (a *Admin) Close() error {
	var lastErr error

	if a.channel != nil {
		if err := a.channel.Close(); err != nil {
			lastErr = err
		}
		a.channel = nil
	}

	if a.conn != nil {
		if err := a.conn.Close(); err != nil {
			lastErr = err
		}
		a.conn = nil
	}

	a.connected = false
	a.config = nil
	return lastErr
}
