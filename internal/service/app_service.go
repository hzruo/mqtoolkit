package service

import (
	"context"
	"fmt"
	"mq-toolkit/internal/database"
	"mq-toolkit/internal/factory"
	"mq-toolkit/internal/logger"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"sync"
	"time"
)

// AppService 应用服务
type AppService struct {
	ctx             context.Context
	db              *database.Database
	logger          *logger.Logger
	configService   *ConfigService
	historyService  *HistoryService
	consumerService *ConsumerService
	templateService *TemplateService
	mqFactory       factory.Factory
	clientsMutex    sync.Mutex
	activeClients   map[string]mq.Client
}

// NewAppService creates a new AppService
func NewAppService(ctx context.Context, db *database.Database, logger *logger.Logger) *AppService {
	configSvc := NewConfigService(db.GetDB())
	historySvc := NewHistoryService(db.GetDB())
	templateSvc := NewTemplateService(db.GetDB())

	appService := &AppService{
		ctx:             ctx,
		db:              db,
		logger:          logger,
		configService:   configSvc,
		historyService:  historySvc,
		templateService: templateSvc,
		activeClients:   make(map[string]mq.Client),
		mqFactory:       factory.NewFactory(),
	}

	appService.consumerService = NewConsumerService(ctx, logger, appService.mqFactory, configSvc, historySvc)

	return appService
}

// GetConfigService 获取配置服务
func (s *AppService) GetConfigService() *ConfigService {
	return s.configService
}

// GetHistoryService 获取历史记录服务
func (s *AppService) GetHistoryService() *HistoryService {
	return s.historyService
}

// GetTemplateService returns the template service
func (s *AppService) GetTemplateService() *TemplateService {
	return s.templateService
}

// TestConnection 测试连接
func (s *AppService) TestConnection(ctx context.Context, connectionID string) *types.TestResult {
	start := time.Now()

	config, err := s.configService.GetConnection(ctx, connectionID)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get connection config: %v", err))
		return &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Failed to get connection config: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}

	admin, err := s.mqFactory.CreateAdmin(config.Type)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to create admin client: %v", err))
		return &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create admin client: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
	}
	defer admin.Close()

	if err := admin.Connect(ctx, config); err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to connect: %v", err))
		result := &types.TestResult{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
			Latency: time.Since(start).Milliseconds(),
		}
		s.historyService.AddTestRecord(ctx, connectionID, false, result.Message, result.Latency)
		return result
	}

	testResult := admin.TestConnection(ctx)
	if testResult == nil {
		testResult = &types.TestResult{
			Success: false,
			Message: "Test returned a nil result",
		}
	}
	testResult.Latency = time.Since(start).Milliseconds()
	s.historyService.AddTestRecord(ctx, connectionID, testResult.Success, testResult.Message, testResult.Latency)
	s.logger.Info("AppService", fmt.Sprintf("Connection test completed for %s: %v", config.Name, testResult.Success))
	return testResult
}

// ProduceMessage 发送消息
func (s *AppService) ProduceMessage(ctx context.Context, req *types.ProduceRequest) error {
	start := time.Now()

	config, err := s.configService.GetConnection(ctx, req.ConnectionID)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get connection config: %v", err))
		return err
	}

	client, err := s.getOrCreateClient(ctx, req.ConnectionID, config)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get producer client: %v", err))
		return err
	}

	err = client.Produce(ctx, req)
	latency := time.Since(start).Milliseconds()

	success := err == nil
	message := "Message sent successfully"
	if err != nil {
		message = fmt.Sprintf("Failed to send message: %v", err)
	}
	s.historyService.AddProduceRecord(ctx, req.ConnectionID, req.Topic, success, message, latency)

	if success {
		s.logger.Info("AppService", fmt.Sprintf("Message sent to topic %s", req.Topic))
	} else {
		s.logger.Error("AppService", fmt.Sprintf("Failed to send message to topic %s: %v", req.Topic, err))
	}

	return err
}

// StartConsuming 调用 ConsumerService 开始消费
func (s *AppService) StartConsuming(req *types.ConsumeRequest) (string, error) {
	s.logger.Info("AppService", fmt.Sprintf("Received request to start consuming from topic(s): %v", req.Topics))
	subscriptionID, err := s.consumerService.StartConsuming(req)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to start consumer: %v", err))
		return "", err
	}
	s.logger.Info("AppService", fmt.Sprintf("Consumer started with subscription ID: %s", subscriptionID))
	return subscriptionID, nil
}

// StopConsuming 调用 ConsumerService 停止消费
func (s *AppService) StopConsuming(subscriptionID string) {
	s.logger.Info("AppService", fmt.Sprintf("Received request to stop subscription: %s", subscriptionID))
	s.consumerService.StopConsuming(subscriptionID)
}

// ListTopics 列出主题
func (s *AppService) ListTopics(ctx context.Context, connectionID string) ([]types.TopicInfo, error) {
	config, err := s.configService.GetConnection(ctx, connectionID)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get connection config: %v", err))
		return nil, err
	}

	client, err := s.getOrCreateClient(ctx, connectionID, config)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get admin client: %v", err))
		return nil, err
	}

	topics, err := client.ListTopics(ctx)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to list topics: %v", err))
		return nil, err
	}

	s.logger.Info("AppService", fmt.Sprintf("Listed %d topics for connection %s", len(topics), connectionID))
	return topics, nil
}

// CreateTopic 创建主题
func (s *AppService) CreateTopic(ctx context.Context, req *types.CreateTopicRequest) error {
	if req.ConnectionID == "" {
		s.logger.Error("AppService", "CreateTopic: ConnectionID is empty")
		return fmt.Errorf("ConnectionID is required")
	}

	config, err := s.configService.GetConnection(ctx, req.ConnectionID)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get connection config: %v", err))
		return err
	}

	client, err := s.getOrCreateClient(ctx, req.ConnectionID, config)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get admin client: %v", err))
		return err
	}

	err = client.CreateTopic(ctx, req.Topic, req.Partitions, req.Replicas)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to create topic %s: %v", req.Topic, err))
		return err
	}

	s.logger.Info("AppService", fmt.Sprintf("Created topic %s with %d partitions and %d replicas", req.Topic, req.Partitions, req.Replicas))
	return nil
}

// DeleteTopic 删除主题
func (s *AppService) DeleteTopic(ctx context.Context, req *types.DeleteTopicRequest) error {
	if req.ConnectionID == "" {
		s.logger.Error("AppService", "DeleteTopic: ConnectionID is empty")
		return fmt.Errorf("ConnectionID is required")
	}

	config, err := s.configService.GetConnection(ctx, req.ConnectionID)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get connection config: %v", err))
		return err
	}

	client, err := s.getOrCreateClient(ctx, req.ConnectionID, config)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to get admin client: %v", err))
		return err
	}

	err = client.DeleteTopic(ctx, req.Topic)
	if err != nil {
		s.logger.Error("AppService", fmt.Sprintf("Failed to delete topic %s: %v", req.Topic, err))
		return err
	}

	s.logger.Info("AppService", fmt.Sprintf("Deleted topic %s", req.Topic))
	return nil
}

// getOrCreateClient 获取或创建客户端 (用于生产者/Admin)
func (s *AppService) getOrCreateClient(ctx context.Context, connectionID string, config *types.ConnectionConfig) (mq.Client, error) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if client, exists := s.activeClients[connectionID]; exists && client.IsConnected() {
		return client, nil
	}

	client, err := s.mqFactory.CreateClient(config.Type)
	if err != nil {
		return nil, err
	}

	if err := client.Connect(ctx, config); err != nil {
		client.Close()
		return nil, err
	}

	s.activeClients[connectionID] = client
	return client, nil
}

// CloseConnection 关闭连接
func (s *AppService) CloseConnection(connectionID string) error {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	if client, exists := s.activeClients[connectionID]; exists {
		err := client.Close()
		delete(s.activeClients, connectionID)
		s.logger.Info("AppService", fmt.Sprintf("Closed connection: %s", connectionID))
		return err
	}
	return nil
}

// Shutdown 关闭应用服务
func (s *AppService) Shutdown() error {
	s.logger.Info("AppService", "Shutting down application service...")
	// 停止所有消费者
	s.consumerService.StopAllConsumers()

	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for connectionID, client := range s.activeClients {
		if err := client.Close(); err != nil {
			s.logger.Error("AppService", fmt.Sprintf("Failed to close connection %s: %v", connectionID, err))
		}
	}

	s.activeClients = make(map[string]mq.Client)
	s.logger.Info("AppService", "Application service shutdown completed")
	return nil
}
