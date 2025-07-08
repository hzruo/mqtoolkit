package rabbitmq

import (
	"context"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
)

// Client RabbitMQ完整客户端实现
type Client struct {
	producer mq.Producer
	consumer mq.Consumer
	admin    mq.Admin
	config   *types.ConnectionConfig
}

// NewClient 创建RabbitMQ完整客户端
func NewClient() mq.Client {
	return &Client{
		producer: NewProducer(),
		consumer: NewConsumer(),
		admin:    NewAdmin(),
	}
}

// Connect 连接到RabbitMQ
func (c *Client) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	c.config = config
	
	// 连接所有组件
	if err := c.producer.Connect(ctx, config); err != nil {
		return err
	}
	
	if err := c.consumer.Connect(ctx, config); err != nil {
		c.producer.Close()
		return err
	}
	
	if err := c.admin.Connect(ctx, config); err != nil {
		c.producer.Close()
		c.consumer.Close()
		return err
	}
	
	return nil
}

// Produce 发送消息
func (c *Client) Produce(ctx context.Context, req *types.ProduceRequest) error {
	return c.producer.Produce(ctx, req)
}

// ProduceBatch 批量发送消息
func (c *Client) ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) error {
	return c.producer.ProduceBatch(ctx, reqs)
}

// Subscribe 订阅队列
func (c *Client) Subscribe(ctx context.Context, req *types.ConsumeRequest) error {
	return c.consumer.Subscribe(ctx, req)
}

// Consume 消费消息
func (c *Client) Consume(ctx context.Context, handler mq.MessageHandler) error {
	return c.consumer.Consume(ctx, handler)
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) *types.TestResult {
	return c.admin.TestConnection(ctx)
}

// ListTopics 列出所有队列
func (c *Client) ListTopics(ctx context.Context) ([]types.TopicInfo, error) {
	return c.admin.ListTopics(ctx)
}

// CreateTopic 创建队列
func (c *Client) CreateTopic(ctx context.Context, topic string, partitions int32, replicas int16) error {
	return c.admin.CreateTopic(ctx, topic, partitions, replicas)
}

// DeleteTopic 删除队列
func (c *Client) DeleteTopic(ctx context.Context, topic string) error {
	return c.admin.DeleteTopic(ctx, topic)
}

// ListConsumerGroups 列出消费组
func (c *Client) ListConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error) {
	return c.admin.ListConsumerGroups(ctx)
}

// Close 关闭客户端
func (c *Client) Close() error {
	var lastErr error
	
	if err := c.producer.Close(); err != nil {
		lastErr = err
	}
	
	if err := c.consumer.Close(); err != nil {
		lastErr = err
	}
	
	if err := c.admin.Close(); err != nil {
		lastErr = err
	}
	
	return lastErr
}

// IsConnected 检查连接状态
func (c *Client) IsConnected() bool {
	return c.producer.IsConnected() && c.consumer.IsConnected()
}
