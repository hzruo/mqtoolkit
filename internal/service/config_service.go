package service

import (
	"context"
	"fmt"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"

	"gorm.io/gorm"
)

// ConfigService 配置管理服务
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建配置服务
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

// CreateConnection 创建连接配置
func (s *ConfigService) CreateConnection(ctx context.Context, config *types.ConnectionConfig) error {
	if config.ID == "" {
		config.ID = utils.GenerateID()
	}
	
	// 检查名称是否重复
	var count int64
	if err := s.db.Model(&types.ConnectionConfig{}).Where("name = ?", config.Name).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check connection name: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("connection name '%s' already exists", config.Name)
	}
	
	return s.db.Create(config).Error
}

// UpdateConnection 更新连接配置
func (s *ConfigService) UpdateConnection(ctx context.Context, config *types.ConnectionConfig) error {
	// 检查连接是否存在
	var existing types.ConnectionConfig
	if err := s.db.First(&existing, "id = ?", config.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("connection not found: %s", config.ID)
		}
		return fmt.Errorf("failed to find connection: %w", err)
	}
	
	// 检查名称是否与其他连接重复
	var count int64
	if err := s.db.Model(&types.ConnectionConfig{}).Where("name = ? AND id != ?", config.Name, config.ID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check connection name: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("connection name '%s' already exists", config.Name)
	}
	
	return s.db.Save(config).Error
}

// DeleteConnection 删除连接配置
func (s *ConfigService) DeleteConnection(ctx context.Context, id string) error {
	result := s.db.Delete(&types.ConnectionConfig{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete connection: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("connection not found: %s", id)
	}
	return nil
}

// GetConnection 获取连接配置
func (s *ConfigService) GetConnection(ctx context.Context, id string) (*types.ConnectionConfig, error) {
	var config types.ConnectionConfig
	if err := s.db.First(&config, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("connection not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	return &config, nil
}

// ListConnections 列出所有连接配置
func (s *ConfigService) ListConnections(ctx context.Context) ([]*types.ConnectionConfig, error) {
	var configs []*types.ConnectionConfig
	if err := s.db.Order("created DESC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}
	return configs, nil
}

// ListConnectionsByType 按类型列出连接配置
func (s *ConfigService) ListConnectionsByType(ctx context.Context, mqType types.MQType) ([]*types.ConnectionConfig, error) {
	var configs []*types.ConnectionConfig
	if err := s.db.Where("type = ?", mqType).Order("created DESC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list connections by type: %w", err)
	}
	return configs, nil
}
