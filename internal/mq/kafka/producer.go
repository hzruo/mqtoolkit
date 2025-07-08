package kafka

import (
	"context"
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer Kafka生产者实现
type Producer struct {
	writer    *kafka.Writer
	connected bool
}

// NewProducer 创建Kafka生产者
func NewProducer() mq.Producer {
	return &Producer{}
}

// Connect 连接到Kafka
func (p *Producer) Connect(ctx context.Context, config *types.ConnectionConfig) error {
	if config.Type != types.MQTypeKafka {
		return utils.NewValidationError("Invalid MQ type for Kafka producer", string(config.Type))
	}

	// 构建broker地址
	brokers := []string{fmt.Sprintf("%s:%d", config.Host, config.Port)}

	// 从Extra配置中获取额外的broker
	if extraBrokers, ok := config.Extra["brokers"]; ok {
		for _, broker := range parseCommaSeparated(extraBrokers) {
			if broker != "" {
				brokers = append(brokers, broker)
			}
		}
	}

	// 创建Writer配置
	writerConfig := kafka.WriterConfig{
		Brokers: brokers,
		// Topic会在发送消息时指定
		Balancer:     &kafka.LeastBytes{}, // 默认使用最少字节负载均衡
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	}

	// 配置认证
	if config.Username != "" && config.Password != "" {
		// 这里需要根据实际情况配置SASL认证
		// kafka-go库的SASL配置比较复杂，这里简化处理
	}

	// 从Extra配置中获取其他参数
	if batchSize, ok := config.Extra["batch_size"]; ok {
		if size, err := strconv.Atoi(batchSize); err == nil && size > 0 {
			writerConfig.BatchSize = size
		}
	}

	if timeout, ok := config.Extra["batch_timeout"]; ok {
		if duration, err := time.ParseDuration(timeout); err == nil {
			writerConfig.BatchTimeout = duration
		}
	}

	p.writer = kafka.NewWriter(writerConfig)
	p.connected = true

	return nil
}

// Produce 发送单条消息
func (p *Producer) Produce(ctx context.Context, req *types.ProduceRequest) error {
	if !p.connected || p.writer == nil {
		return utils.NewConnectionError("Producer not connected", nil)
	}

	if !utils.IsValidTopic(req.Topic) {
		return utils.NewValidationError("Invalid topic name", req.Topic)
	}

	// 构建Kafka消息
	message := kafka.Message{
		Topic: req.Topic,
		Key:   []byte(req.Key),
		Value: []byte(req.Value),
		Time:  time.Now(),
	}

	// 添加Headers
	if req.Headers != nil {
		headers := make([]kafka.Header, 0, len(req.Headers))
		for key, value := range req.Headers {
			headers = append(headers, kafka.Header{
				Key:   key,
				Value: []byte(value),
			})
		}
		message.Headers = headers
	}

	// 指定分区（如果提供）
	if req.Partition != nil {
		message.Partition = int(*req.Partition)
	}

	// 发送消息
	return p.writer.WriteMessages(ctx, message)
}

// ProduceBatch 批量发送消息
func (p *Producer) ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) error {
	if !p.connected || p.writer == nil {
		return utils.NewConnectionError("Producer not connected", nil)
	}

	if len(reqs) == 0 {
		return utils.NewValidationError("Empty message batch", "")
	}

	// 构建Kafka消息批次
	messages := make([]kafka.Message, 0, len(reqs))
	for _, req := range reqs {
		if !utils.IsValidTopic(req.Topic) {
			return utils.NewValidationError("Invalid topic name", req.Topic)
		}

		message := kafka.Message{
			Topic: req.Topic,
			Key:   []byte(req.Key),
			Value: []byte(req.Value),
			Time:  time.Now(),
		}

		// 添加Headers
		if req.Headers != nil {
			headers := make([]kafka.Header, 0, len(req.Headers))
			for key, value := range req.Headers {
				headers = append(headers, kafka.Header{
					Key:   key,
					Value: []byte(value),
				})
			}
			message.Headers = headers
		}

		// 指定分区（如果提供）
		if req.Partition != nil {
			message.Partition = int(*req.Partition)
		}

		messages = append(messages, message)
	}

	// 批量发送消息
	return p.writer.WriteMessages(ctx, messages...)
}

// Close 关闭生产者
func (p *Producer) Close() error {
	if p.writer != nil {
		err := p.writer.Close()
		p.writer = nil
		p.connected = false
		return err
	}
	return nil
}

// IsConnected 检查连接状态
func (p *Producer) IsConnected() bool {
	return p.connected && p.writer != nil
}
