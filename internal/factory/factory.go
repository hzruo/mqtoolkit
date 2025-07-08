package factory

import (
	"fmt"
	"mq-toolkit/internal/mq"
	"mq-toolkit/internal/mq/kafka"
	"mq-toolkit/internal/mq/rabbitmq"
	"mq-toolkit/internal/mq/rocketmq"
	"mq-toolkit/pkg/types"
)

// Factory defines the interface for creating MQ clients
type Factory interface {
	CreateClient(mqType types.MQType) (mq.Client, error)
	CreateAdmin(mqType types.MQType) (mq.Admin, error)
	CreateProducer(mqType types.MQType) (mq.Producer, error)
	CreateConsumer(mqType types.MQType) (mq.Consumer, error)
}

// NewFactory creates a new MQ client factory
func NewFactory() Factory {
	return &factory{}
}

type factory struct{}

func (f *factory) CreateClient(mqType types.MQType) (mq.Client, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewClient(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewClient(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewClient(), nil
	default:
		return nil, fmt.Errorf("unsupported mq type: %s", mqType)
	}
}

func (f *factory) CreateAdmin(mqType types.MQType) (mq.Admin, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewAdmin(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewAdmin(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewAdmin(), nil
	default:
		return nil, fmt.Errorf("unsupported mq type: %s", mqType)
	}
}

func (f *factory) CreateProducer(mqType types.MQType) (mq.Producer, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewProducer(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewProducer(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewProducer(), nil
	default:
		return nil, fmt.Errorf("unsupported mq type: %s", mqType)
	}
}

func (f *factory) CreateConsumer(mqType types.MQType) (mq.Consumer, error) {
	switch mqType {
	case types.MQTypeKafka:
		return kafka.NewConsumer(), nil
	case types.MQTypeRabbitMQ:
		return rabbitmq.NewConsumer(), nil
	case types.MQTypeRocketMQ:
		return rocketmq.NewConsumer(), nil
	default:
		return nil, fmt.Errorf("unsupported mq type: %s", mqType)
	}
}
