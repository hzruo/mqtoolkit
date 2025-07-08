package kafka

import (
	"context"
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Consumer Kafka消费者实现
type Consumer struct {
	reader    *kafka.Reader
	connected bool
	config    *types.ConnectionConfig
}

// NewConsumer 创建Kafka消费者
func NewConsumer() mq.Consumer {
	return &Consumer{}
}

// Connect 连接到Kafka
func (c *Consumer) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeKafka {
		return utils.NewValidationError("Invalid MQ type for Kafka consumer", string(config.Type))
	}
	// 保存配置以备后用
	c.config = config
	c.connected = true
	return nil
}

// Subscribe 订阅主题并准备消费
func (c *Consumer) Subscribe(ctx context.Context, req *types.ConsumeRequest) error {
	if !c.connected || c.config == nil {
		return utils.NewConnectionError("Consumer not connected", nil)
	}

	if len(req.Topics) == 0 {
		return utils.NewValidationError("No topics specified for subscription", "")
	}

	// 构建broker地址
	brokers := []string{fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)}
	if extraBrokers, ok := c.config.Extra["brokers"]; ok {
		for _, broker := range parseCommaSeparated(extraBrokers) {
			if broker != "" {
				brokers = append(brokers, broker)
			}
		}
	}

	// 确定消费组ID
	groupID := req.GroupID
	if groupID == "" {
		groupID = c.config.GroupID
	}
	if groupID == "" {
		groupID = "mq-toolkit-default-consumer"
	}

	// 创建Reader配置 - kafka-go v0.4.48只支持单个Topic
	// 如果有多个topics，我们只使用第一个
	topic := ""
	if len(req.Topics) > 0 {
		topic = req.Topics[0]
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MaxBytes: 10e6, // 10MB
	}

	// 配置起始位置
	if req.FromBeginning {
		readerConfig.StartOffset = kafka.FirstOffset
	} else {
		readerConfig.StartOffset = kafka.LastOffset
	}

	// 从Extra配置中获取其他参数
	if maxBytes, ok := c.config.Extra["max_bytes"]; ok {
		if size, err := strconv.Atoi(maxBytes); err == nil && size > 0 {
			readerConfig.MaxBytes = size
		}
	}

	if minBytes, ok := c.config.Extra["min_bytes"]; ok {
		if size, err := strconv.Atoi(minBytes); err == nil && size > 0 {
			readerConfig.MinBytes = size
		}
	}

	// 创建一个新的Reader
	c.reader = kafka.NewReader(readerConfig)

	return nil
}

// Consume 消费消息
func (c *Consumer) Consume(ctx context.Context, handler mq.MessageHandler) error {
	if !c.connected {
		return utils.NewConnectionError("Consumer not connected", nil)
	}

	if c.reader == nil {
		return utils.NewConnectionError("Consumer not subscribed to any topic. Call Subscribe first.", nil)
	}

	// 开始消费消息
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 读取消息
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				// 如果是上下文取消，则正常退出
				if err == context.Canceled || err == context.DeadlineExceeded {
					return nil
				}
				// 检查是否是连接关闭错误（通常在停止消费时发生）
				if strings.Contains(err.Error(), "connection closed") ||
					strings.Contains(err.Error(), "use of closed network connection") ||
					strings.Contains(err.Error(), "CONN_") {
					return nil // 正常退出，不报告错误
				}
				return utils.NewConnectionError("Failed to read message", err)
			}

			// 转换为内部消息格式
			msg := &types.Message{
				ID:        utils.GenerateID(),
				Topic:     message.Topic,
				Key:       string(message.Key),
				Value:     string(message.Value),
				Partition: int32(message.Partition),
				Offset:    message.Offset,
				Timestamp: message.Time,
			}

			// 转换Headers
			if len(message.Headers) > 0 {
				msg.Headers = make(map[string]string)
				for _, header := range message.Headers {
					msg.Headers[header.Key] = string(header.Value)
				}
			}

			// 调用处理器
			if err := handler(msg); err != nil {
				// 记录错误但继续处理，除非处理器返回特定错误
				fmt.Printf("message handler error: %v\n", err)
			}
		}
	}
}

// Close 关闭消费者
func (c *Consumer) Close() error {
	c.connected = false
	c.config = nil
	if c.reader != nil {
		err := c.reader.Close()
		c.reader = nil
		return err
	}
	return nil
}

// IsConnected 检查连接状态
func (c *Consumer) IsConnected() bool {
	return c.connected
}
