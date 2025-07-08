package kafka

import (
	"context"
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"net"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// Admin Kafka管理客户端实现
type Admin struct {
	conn      *kafka.Conn
	connected bool
	config    *types.ConnectionConfig
}

// NewAdmin 创建Kafka管理客户端
func NewAdmin() mq.Admin {
	return &Admin{}
}

// Connect 连接到Kafka
func (a *Admin) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeKafka {
		return utils.NewValidationError("Invalid MQ type for Kafka admin", string(config.Type))
	}

	// 保存配置
	a.config = config

	// 建立连接
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		return utils.NewConnectionError("Failed to connect to Kafka", err)
	}

	a.conn = conn
	a.connected = true
	return nil
}

// TestConnection 测试连接
func (a *Admin) TestConnection(ctx context.Context) *types.TestResult {
	start := time.Now()

	if !a.connected || a.conn == nil {
		return &types.TestResult{
			Success: false,
			Message: "Not connected to Kafka",
			Latency: time.Since(start).Milliseconds(),
		}
	}

	// 尝试获取broker信息
	brokers, err := a.conn.Brokers()
	if err != nil {
		return &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get brokers: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	return &types.TestResult{
		Success: true,
		Message: fmt.Sprintf("Connected successfully. Found %d brokers", len(brokers)),
		Latency: time.Since(start).Milliseconds(),
	}
}

// ListTopics 列出所有主题
func (a *Admin) ListTopics(ctx context.Context) ([]types.TopicInfo, error) {
	if !a.connected || a.conn == nil {
		return nil, utils.NewConnectionError("Not connected to Kafka", nil)
	}

	// 重新建立连接以确保连接有效
	address := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		return nil, utils.NewConnectionError("Failed to reconnect to Kafka", err)
	}
	defer conn.Close()

	// 获取分区信息
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return nil, utils.NewConnectionError("Failed to read partitions", err)
	}

	// 按主题分组
	topicMap := make(map[string]*types.TopicInfo)
	for _, partition := range partitions {
		if topic, exists := topicMap[partition.Topic]; exists {
			if int32(partition.ID) >= topic.Partitions {
				topic.Partitions = int32(partition.ID) + 1
			}
		} else {
			topicMap[partition.Topic] = &types.TopicInfo{
				Name:       partition.Topic,
				Partitions: int32(partition.ID) + 1,
				Replicas:   int16(len(partition.Replicas)),
			}
		}
	}

	// 转换为切片
	var topics []types.TopicInfo
	for _, topic := range topicMap {
		topics = append(topics, *topic)
	}

	return topics, nil
}

// CreateTopic 创建主题
func (a *Admin) CreateTopic(ctx context.Context, topic string, partitions int32, replicas int16) error {
	if !a.connected || a.conn == nil {
		return utils.NewConnectionError("Not connected to Kafka", nil)
	}

	if !utils.IsValidTopic(topic) {
		return utils.NewValidationError("Invalid topic name", topic)
	}

	// 重新建立连接以确保连接有效
	address := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		return utils.NewConnectionError("Failed to reconnect to Kafka", err)
	}
	defer conn.Close()

	// 获取控制器连接
	controller, err := conn.Controller()
	if err != nil {
		return utils.NewConnectionError("Failed to get controller", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return utils.NewConnectionError("Failed to connect to controller", err)
	}
	defer controllerConn.Close()

	// 创建主题配置
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     int(partitions),
			ReplicationFactor: int(replicas),
		},
	}

	// 创建主题
	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return utils.NewConnectionError("Failed to create topic", err)
	}

	return nil
}

// DeleteTopic 删除主题
func (a *Admin) DeleteTopic(ctx context.Context, topic string) error {
	if !a.connected || a.conn == nil {
		return utils.NewConnectionError("Not connected to Kafka", nil)
	}

	if !utils.IsValidTopic(topic) {
		return utils.NewValidationError("Invalid topic name", topic)
	}

	// 重新建立连接以确保连接有效
	address := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)
	conn, err := kafka.Dial("tcp", address)
	if err != nil {
		return utils.NewConnectionError("Failed to reconnect to Kafka", err)
	}
	defer conn.Close()

	// 获取控制器连接
	controller, err := conn.Controller()
	if err != nil {
		return utils.NewConnectionError("Failed to get controller", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return utils.NewConnectionError("Failed to connect to controller", err)
	}
	defer controllerConn.Close()

	// 删除主题
	err = controllerConn.DeleteTopics(topic)
	if err != nil {
		return utils.NewConnectionError("Failed to delete topic", err)
	}

	return nil
}

// ListConsumerGroups 列出消费组
func (a *Admin) ListConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error) {
	if !a.connected || a.conn == nil {
		return nil, utils.NewConnectionError("Not connected to Kafka", nil)
	}

	// kafka-go库没有直接的API来列出消费组
	// 这里返回空列表，实际实现需要使用更高级的API
	return []types.ConsumerGroup{}, nil
}

// Close 关闭连接
func (a *Admin) Close() error {
	if a.conn != nil {
		err := a.conn.Close()
		a.conn = nil
		a.connected = false
		a.config = nil
		return err
	}
	return nil
}
