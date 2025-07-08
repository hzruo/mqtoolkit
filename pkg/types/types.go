package types

import (
	"context"
	"time"
)

// MQType 消息队列类型
type MQType string

const (
	MQTypeKafka    MQType = "kafka"
	MQTypeRabbitMQ MQType = "rabbitmq"
	MQTypeRocketMQ MQType = "rocketmq"
)

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	ID       string            `json:"id" gorm:"primaryKey"`
	Name     string            `json:"name" gorm:"not null"`
	Type     MQType            `json:"type" gorm:"not null"`
	Host     string            `json:"host" gorm:"not null"`
	Port     int               `json:"port" gorm:"not null"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	VHost    string            `json:"vhost"`    // RabbitMQ virtual host
	GroupID  string            `json:"group_id"` // Kafka consumer group
	Extra    map[string]string `json:"extra" gorm:"serializer:json"`
	Created  time.Time         `json:"created" gorm:"autoCreateTime"`
	Updated  time.Time         `json:"updated" gorm:"autoUpdateTime"`
}

// Message 消息结构
type Message struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Headers   map[string]string `json:"headers"`
	Partition int32             `json:"partition"`
	Offset    int64             `json:"offset"`
	Timestamp time.Time         `json:"timestamp"`
}

// ProduceRequest 生产消息请求
type ProduceRequest struct {
	ConnectionID string            `json:"connection_id"`
	Topic        string            `json:"topic"`
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Headers      map[string]string `json:"headers"`
	Partition    *int32            `json:"partition,omitempty"`
}

// ConsumeRequest 消费消息请求
type ConsumeRequest struct {
	ConnectionID  string   `json:"connection_id"`
	Topics        []string `json:"topics"`
	GroupID       string   `json:"group_id"`
	AutoCommit    bool     `json:"auto_commit"`
	FromBeginning bool     `json:"from_beginning"`
}

// TestResult 测试结果
type TestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency"` // 延迟（毫秒）
}

// HistoryRecord represents a single entry in the operation history
type HistoryRecord struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	ConnectionID string    `json:"connection_id"`
	Type         string    `json:"type"` // "produce", "consume", "test_connection"
	Topic        string    `json:"topic"`
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	Latency      int64     `json:"latency"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created"`
}

// MessageTemplate represents a reusable message template
type MessageTemplate struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName for HistoryRecord
func (HistoryRecord) TableName() string {
	return "history_records"
}

// TableName for MessageTemplate
func (MessageTemplate) TableName() string {
	return "message_templates"
}

// LogEntry 日志条目
type LogEntry struct {
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// TopicInfo 主题信息
type TopicInfo struct {
	Name       string `json:"name"`
	Partitions int32  `json:"partitions"`
	Replicas   int16  `json:"replicas"`
}

// ConsumerGroup 消费组信息
type ConsumerGroup struct {
	ID      string   `json:"id"`
	Members []string `json:"members"`
	Topics  []string `json:"topics"`
}

// CreateTopicRequest 创建主题请求
type CreateTopicRequest struct {
	ConnectionID string `json:"connection_id"`
	Topic        string `json:"topic"`
	Partitions   int32  `json:"partitions"`
	Replicas     int16  `json:"replicas"`
}

// ... (previous content) ...

// DeleteTopicRequest 删除主题请求
type DeleteTopicRequest struct {
	ConnectionID string `json:"connection_id"`
	Topic        string `json:"topic"`
}

// ConfigService defines the interface for the configuration service
type ConfigService interface {
	GetConnection(ctx context.Context, id string) (*ConnectionConfig, error)
	ListConnections(ctx context.Context) ([]*ConnectionConfig, error)
	CreateConnection(ctx context.Context, config *ConnectionConfig) error
	UpdateConnection(ctx context.Context, config *ConnectionConfig) error
	DeleteConnection(ctx context.Context, id string) error
}

// HistoryService defines the interface for the history service
type HistoryService interface {
	AddRecord(ctx context.Context, record *HistoryRecord) error
	AddProduceRecord(ctx context.Context, connectionID, topic string, success bool, message string, latency int64) error
	AddConsumeRecord(ctx context.Context, connectionID, topic string, success bool, message string, latency int64) error
	AddTestRecord(ctx context.Context, connectionID string, success bool, message string, latency int64) error
	GetRecords(ctx context.Context, limit, offset int) ([]*HistoryRecord, error)
	ClearRecords(ctx context.Context) error
}
