package service

import (
	"context"
	"fmt"
	"mq-toolkit/internal/factory"
	"mq-toolkit/internal/logger"
	"mq-toolkit/internal/mq"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ConsumerService 负责管理消息消费
type ConsumerService struct {
	ctx        context.Context
	logger     *logger.Logger
	mqFactory  factory.Factory
	configSvc  *ConfigService
	historySvc *HistoryService
	activeSubs sync.Map // 存储活跃的订阅 [subscriptionID -> *activeSubscription]
}

// activeSubscription 代表一个活跃的订阅
type activeSubscription struct {
	consumer     mq.Consumer
	cancel       context.CancelFunc
	connectionID string
	topics       []string
}

// NewConsumerService 创建一个新的 ConsumerService
func NewConsumerService(ctx context.Context, logger *logger.Logger, factory factory.Factory, configSvc *ConfigService, historySvc *HistoryService) *ConsumerService {
	return &ConsumerService{
		ctx:        ctx,
		logger:     logger,
		mqFactory:  factory,
		configSvc:  configSvc,
		historySvc: historySvc,
	}
}

// StartConsuming 开始消费消息
func (s *ConsumerService) StartConsuming(req *types.ConsumeRequest) (string, error) {
	// 获取连接配置
	connConfig, err := s.configSvc.GetConnection(s.ctx, req.ConnectionID)
	if err != nil {
		return "", fmt.Errorf("failed to get connection config: %w", err)
	}

	// 创建消费者
	consumer, err := s.mqFactory.CreateConsumer(connConfig.Type)
	if err != nil {
		return "", fmt.Errorf("failed to create consumer: %w", err)
	}

	// 连接
	if err := consumer.Connect(s.ctx, connConfig); err != nil {
		return "", fmt.Errorf("failed to connect: %w", err)
	}

	// 订阅
	if err := consumer.Subscribe(s.ctx, req); err != nil {
		consumer.Close()
		return "", fmt.Errorf("failed to subscribe: %w", err)
	}

	// 创建一个可取消的上下文来控制消费 goroutine
	consumeCtx, cancel := context.WithCancel(s.ctx)
	subscriptionID := utils.GenerateID()

	// 保存订阅信息
	s.activeSubs.Store(subscriptionID, &activeSubscription{
		consumer:     consumer,
		cancel:       cancel,
		connectionID: req.ConnectionID,
		topics:       req.Topics,
	})

	// 在一个新的 goroutine 中开始消费
	go func() {
		s.logger.Info("ConsumerService", fmt.Sprintf("Starting consumer for subscription %s", subscriptionID))

		err := consumer.Consume(consumeCtx, func(msg *types.Message) error {
			// 记录收到的消息
			s.logger.Info("ConsumerService", fmt.Sprintf("Received message from topic %s: %s", msg.Topic, msg.Value))

			// 记录消费历史
			s.historySvc.AddConsumeRecord(s.ctx, req.ConnectionID, msg.Topic, true, fmt.Sprintf("Consumed message: %s", msg.Value), 0)

			// 将消息发送到前端
			runtime.EventsEmit(s.ctx, "message:received", msg)
			s.logger.Info("ConsumerService", "Message event emitted to frontend")
			return nil
		})

		// 消费结束后（可能因为错误或取消）
		if err != nil && err != context.Canceled {
			// 检查是否是正常的连接关闭错误
			errMsg := err.Error()
			isNormalShutdown := strings.Contains(errMsg, "connection closed") ||
				strings.Contains(errMsg, "use of closed network connection") ||
				strings.Contains(errMsg, "CONN_")

			if !isNormalShutdown {
				s.logger.Error("ConsumerService", fmt.Sprintf("Consumption error for %s: %v", subscriptionID, err))
				runtime.EventsEmit(s.ctx, "consumer:error", map[string]string{
					"subscriptionId": subscriptionID,
					"error":          err.Error(),
				})
			} else {
				s.logger.Info("ConsumerService", fmt.Sprintf("Consumer %s stopped normally", subscriptionID))
			}
		}

		s.logger.Info("ConsumerService", fmt.Sprintf("Stopping consumer for subscription %s", subscriptionID))
		s.stopAndRemove(subscriptionID)
	}()

	return subscriptionID, nil
}

// StopConsuming 停止消费消息
func (s *ConsumerService) StopConsuming(subscriptionID string) {
	s.stopAndRemove(subscriptionID)
}

// stopAndRemove 停止并移除一个订阅
func (s *ConsumerService) stopAndRemove(subscriptionID string) {
	if sub, ok := s.activeSubs.Load(subscriptionID); ok {
		activeSub := sub.(*activeSubscription)
		activeSub.cancel()         // 取消上下文，这将导致 Consume 循环退出
		activeSub.consumer.Close() // 关闭底层连接
		s.activeSubs.Delete(subscriptionID)
		s.logger.Info("ConsumerService", fmt.Sprintf("Successfully stopped and removed subscription %s", subscriptionID))
	}
}

// StopAllConsumers 关闭所有活跃的消费者
func (s *ConsumerService) StopAllConsumers() {
	s.activeSubs.Range(func(key, value interface{}) bool {
		s.StopConsuming(key.(string))
		return true
	})
}
