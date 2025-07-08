package main

import (
	"context"
	"fmt"
	"mq-toolkit/internal/config"
	"mq-toolkit/internal/database"
	"mq-toolkit/internal/logger"
	"mq-toolkit/internal/service"
	"mq-toolkit/pkg/types"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	db         *database.Database
	logger     *logger.Logger
	appService *service.AppService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.logger = logger.NewDefault()
	a.logger.Info("App", "Application starting...")

	// Add a listener to forward logs to the frontend
	a.logger.AddListener(func(entry types.LogEntry) {
		runtime.EventsEmit(a.ctx, "log:new", entry)
	})

	cfg, err := config.Load("config/app.json")
	if err != nil {
		a.logger.Error("App", fmt.Sprintf("Failed to load config: %v", err))
		cfg = &config.Config{
			Database: config.DatabaseConfig{
				Path: "data/mq-toolkit.db",
			},
		}
		a.logger.Info("App", "Using default configuration")
	}

	a.db, err = database.New(cfg.Database.Path)
	if err != nil {
		a.logger.Error("App", fmt.Sprintf("Failed to initialize database: %v", err))
		a.db, err = database.New(":memory:")
		if err != nil {
			a.logger.Error("App", fmt.Sprintf("Failed to initialize memory database: %v", err))
			return
		}
		a.logger.Info("App", "Using in-memory database")
	}

	// 初始化应用服务, 传入ctx
	a.appService = service.NewAppService(a.ctx, a.db, a.logger)

	a.logger.Info("App", "Application started successfully")
}

// shutdown is called when the app terminates
func (a *App) shutdown(ctx context.Context) {
	a.logger.Info("App", "Application shutting down...")
	if a.appService != nil {
		a.appService.Shutdown()
	}
	a.logger.Info("App", "Application shutdown complete.")
}

// GetConnections 获取所有连接配置
func (a *App) GetConnections() ([]*types.ConnectionConfig, error) {
	return a.appService.GetConfigService().ListConnections(a.ctx)
}

// CreateConnection 创建连接配置
func (a *App) CreateConnection(config *types.ConnectionConfig) error {
	return a.appService.GetConfigService().CreateConnection(a.ctx, config)
}

// UpdateConnection 更新连接配置
func (a *App) UpdateConnection(config *types.ConnectionConfig) error {
	return a.appService.GetConfigService().UpdateConnection(a.ctx, config)
}

// DeleteConnection 删除连接配置
func (a *App) DeleteConnection(connectionID string) error {
	return a.appService.GetConfigService().DeleteConnection(a.ctx, connectionID)
}

// TestConnection 测试连接
func (a *App) TestConnection(connectionID string) (res *types.TestResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			a.logger.Error("App", fmt.Sprintf("Panic recovered in TestConnection: %v", r))
			// 创建一个新的错误返回
			err = fmt.Errorf("a panic occurred during connection test: %v", r)
			// 确保我们不会返回一个 nil 的结果, 这会导致JSON错误
			res = &types.TestResult{
				Success: false,
				Message: err.Error(),
			}
		}
	}()

	res = a.appService.TestConnection(a.ctx, connectionID)
	return res, nil // 在成功时显式返回 nil 错误
}

// ProduceMessage 发送消息
func (a *App) ProduceMessage(req *types.ProduceRequest) error {
	return a.appService.ProduceMessage(a.ctx, req)
}

// StartConsuming 开始消费消息
func (a *App) StartConsuming(req *types.ConsumeRequest) (string, error) {
	return a.appService.StartConsuming(req)
}

// StopConsuming 停止消费消息
func (a *App) StopConsuming(subscriptionID string) {
	a.appService.StopConsuming(subscriptionID)
}

// GetHistory 获取历史记录
func (a *App) GetHistory(limit, offset int) ([]*types.HistoryRecord, error) {
	return a.appService.GetHistoryService().GetRecords(a.ctx, limit, offset)
}

// ClearHistory 清空历史记录
func (a *App) ClearHistory() error {
	return a.appService.GetHistoryService().ClearRecords(a.ctx)
}

// GetLogs 获取当前日志
func (a *App) GetLogs() []types.LogEntry {
	return a.logger.GetEntries()
}

// ListTemplates returns all message templates
func (a *App) ListTemplates() ([]*types.MessageTemplate, error) {
	return a.appService.GetTemplateService().ListTemplates(a.ctx)
}

// CreateTemplate creates a new message template
func (a *App) CreateTemplate(name, content string) (*types.MessageTemplate, error) {
	return a.appService.GetTemplateService().CreateTemplate(a.ctx, name, content)
}

// UpdateTemplate updates an existing message template
func (a *App) UpdateTemplate(id, name, content string) error {
	return a.appService.GetTemplateService().UpdateTemplate(a.ctx, id, name, content)
}

// DeleteTemplate deletes a message template
func (a *App) DeleteTemplate(id string) error {
	return a.appService.GetTemplateService().DeleteTemplate(a.ctx, id)
}

// ListTopics 列出主题
func (a *App) ListTopics(connectionID string) ([]types.TopicInfo, error) {
	return a.appService.ListTopics(a.ctx, connectionID)
}

// CreateTopic 创建主题
func (a *App) CreateTopic(req *types.CreateTopicRequest) error {
	return a.appService.CreateTopic(a.ctx, req)
}

// DeleteTopic 删除主题
func (a *App) DeleteTopic(req *types.DeleteTopicRequest) error {
	return a.appService.DeleteTopic(a.ctx, req)
}

// SaveFile 保存文件，支持用户选择路径
func (a *App) SaveFile(filename, content string) (string, error) {
	// 添加mq-toolkit前缀
	prefixedFilename := "mq-toolkit-" + filename

	// 使用Wails的文件保存对话框
	options := runtime.SaveDialogOptions{
		DefaultFilename: prefixedFilename,
		Title:           "保存文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON文件 (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "文本文件 (*.txt)",
				Pattern:     "*.txt",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	}

	// 显示保存对话框
	selectedPath, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", fmt.Errorf("failed to show save dialog: %w", err)
	}

	// 用户取消了保存
	if selectedPath == "" {
		return "", fmt.Errorf("save cancelled by user")
	}

	// 写入文件
	if err := os.WriteFile(selectedPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	a.logger.Info("App", fmt.Sprintf("File saved to: %s", selectedPath))
	return selectedPath, nil
}
