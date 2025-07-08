package service

import (
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/internal/mq/kafka"
	"mq-toolkit/internal/mq/rabbitmq"
	"mq-toolkit/internal/mq/rocketmq"
	"mq-toolkit/pkg/types"
)

// MQFactory 消息队列工厂实现
type MQFactory struct{}

// NewMQFactory 创建新的工厂实例
func NewMQFactory() *MQFactory {
	return &MQFactory{}
}

// CreateProducer 创建生产者
func (f *MQFactory) CreateProducer(mqType types.MQType) (mq.Producer, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewProducer(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewProducer(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewProducer(), nil
	default:
		return nil, fmt.Errorf("unsupported MQ type: %s", mqType)
	}
}

// CreateConsumer 创建消费者
func (f *MQFactory) CreateConsumer(mqType types.MQType) (mq.Consumer, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewConsumer(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewConsumer(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewConsumer(), nil
	default:
		return nil, fmt.Errorf("unsupported MQ type: %s", mqType)
	}
}

// CreateAdmin 创建管理客户端
func (f *MQFactory) CreateAdmin(mqType types.MQType) (mq.Admin, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewAdmin(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewAdmin(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewAdmin(), nil
	default:
		return nil, fmt.Errorf("unsupported MQ type: %s", mqType)
	}
}

// CreateClient 创建完整客户端
func (f *MQFactory) CreateClient(mqType types.MQType) (mq.Client, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewClient(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewClient(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported MQ type: %s", mqType)
	}
}
