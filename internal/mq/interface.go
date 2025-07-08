package mq

import (
	"context"
	"mq-toolkit/pkg/types"
)

// Client is the interface that wraps the basic methods of a message queue client.
type Client interface {
	Connect(ctx context.Context, config *types.ConnectionConfig) error
	Producer
	Consumer
	Admin
	Close() error
	IsConnected() bool
}

// Producer is the interface that wraps the basic methods of a message queue producer.
type Producer interface {
	Connect(ctx context.Context, config *types.ConnectionConfig) error
	Produce(ctx context.Context, req *types.ProduceRequest) error
	ProduceBatch(ctx context.Context, reqs []*types.ProduceRequest) error
	Close() error
	IsConnected() bool
}

// Consumer is the interface that wraps the basic methods of a message queue consumer.
type Consumer interface {
	Connect(ctx context.Context, config *types.ConnectionConfig) error
	Subscribe(ctx context.Context, req *types.ConsumeRequest) error
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
	IsConnected() bool
}

// MessageHandler is the handler for consumed messages.
type MessageHandler func(msg *types.Message) error

// Admin is the interface that wraps the basic methods of a message queue admin client.
type Admin interface {
	Connect(ctx context.Context, config *types.ConnectionConfig) error
	TestConnection(ctx context.Context) *types.TestResult
	ListTopics(ctx context.Context) ([]types.TopicInfo, error)
	CreateTopic(ctx context.Context, topic string, partitions int32, replicas int16) error
	DeleteTopic(ctx context.Context, topic string) error
	ListConsumerGroups(ctx context.Context) ([]types.ConsumerGroup, error)
	Close() error
}
