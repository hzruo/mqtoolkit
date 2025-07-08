package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"net/http"
	"net/url"
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
	// 如果没有提供用户名和密码，使用RabbitMQ默认的guest/guest
	username := config.Username
	password := config.Password
	if username == "" {
		username = "guest"
	}
	if password == "" {
		password = "guest"
	}

	url = fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		username, password, config.Host, config.Port, vhost)

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

// QueueInfo RabbitMQ队列信息结构
type QueueInfo struct {
	Name       string `json:"name"`
	VHost      string `json:"vhost"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto_delete"`
	Messages   int    `json:"messages"`
	Consumers  int    `json:"consumers"`
}

// ListTopics 列出所有队列（RabbitMQ中的"主题"概念对应队列）
func (a *Admin) ListTopics(ctx context.Context) ([]types.TopicInfo, error) {
	if !a.connected || a.channel == nil {
		return nil, utils.NewConnectionError("Not connected to RabbitMQ", nil)
	}

	// 使用RabbitMQ HTTP管理API获取队列列表
	queues, err := a.getQueuesFromAPI()
	if err != nil {
		// 如果HTTP API失败，返回空列表而不是错误，这样不会阻止其他功能
		return []types.TopicInfo{}, nil
	}

	// 转换为TopicInfo格式
	var topics []types.TopicInfo
	for _, queue := range queues {
		topics = append(topics, types.TopicInfo{
			Name:       queue.Name,
			Partitions: 1, // RabbitMQ队列没有分区概念
			Replicas:   1, // RabbitMQ队列没有副本概念
		})
	}

	return topics, nil
}

// getQueuesFromAPI 通过HTTP管理API获取队列列表
func (a *Admin) getQueuesFromAPI() ([]QueueInfo, error) {
	// 构建管理API URL
	vhost := a.config.VHost
	if vhost == "" {
		vhost = "/"
	}

	// URL编码vhost
	encodedVhost := url.QueryEscape(vhost)

	// 管理API默认端口是15672
	managementPort := 15672
	apiURL := fmt.Sprintf("http://%s:%d/api/queues/%s", a.config.Host, managementPort, encodedVhost)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置认证
	username := a.config.Username
	password := a.config.Password
	if username == "" {
		username = "guest"
	}
	if password == "" {
		password = "guest"
	}
	req.SetBasicAuth(username, password)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call management API: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("management API returned status %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析JSON响应
	var queues []QueueInfo
	if err := json.Unmarshal(body, &queues); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return queues, nil
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
