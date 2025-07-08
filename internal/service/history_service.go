package service

import (
	"context"
	"fmt"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"
	"time"

	"gorm.io/gorm"
)

// HistoryService 历史记录管理服务
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService 创建历史记录服务
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// AddRecord 添加历史记录
func (s *HistoryService) AddRecord(ctx context.Context, record *types.HistoryRecord) error {
	if record.ID == "" {
		record.ID = utils.GenerateID()
	}
	return s.db.Create(record).Error
}

// AddProduceRecord 添加生产记录
func (s *HistoryService) AddProduceRecord(ctx context.Context, connectionID, topic string, success bool, message string, latency int64) error {
	record := &types.HistoryRecord{
		ID:           utils.GenerateID(),
		ConnectionID: connectionID,
		Type:         "produce",
		Topic:        topic,
		Success:      success,
		Message:      message,
		Latency:      latency,
	}
	return s.AddRecord(ctx, record)
}

// AddConsumeRecord 添加消费记录
func (s *HistoryService) AddConsumeRecord(ctx context.Context, connectionID, topic string, success bool, message string, latency int64) error {
	record := &types.HistoryRecord{
		ID:           utils.GenerateID(),
		ConnectionID: connectionID,
		Type:         "consume",
		Topic:        topic,
		Success:      success,
		Message:      message,
		Latency:      latency,
	}
	return s.AddRecord(ctx, record)
}

// AddTestRecord 添加测试记录
func (s *HistoryService) AddTestRecord(ctx context.Context, connectionID string, success bool, message string, latency int64) error {
	record := &types.HistoryRecord{
		ID:           utils.GenerateID(),
		ConnectionID: connectionID,
		Type:         "test_connection",
		Topic:        "", // 测试记录没有主题
		Success:      success,
		Message:      message,
		Latency:      latency,
	}
	return s.AddRecord(ctx, record)
}

// GetRecords 获取历史记录
func (s *HistoryService) GetRecords(ctx context.Context, limit, offset int) ([]*types.HistoryRecord, error) {
	var records []*types.HistoryRecord
	query := s.db.Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get records: %w", err)
	}
	return records, nil
}

// GetRecordsByConnection 按连接获取历史记录
func (s *HistoryService) GetRecordsByConnection(ctx context.Context, connectionID string, limit, offset int) ([]*types.HistoryRecord, error) {
	var records []*types.HistoryRecord
	query := s.db.Where("connection_id = ?", connectionID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get records by connection: %w", err)
	}
	return records, nil
}

// GetRecordsByType 按类型获取历史记录
func (s *HistoryService) GetRecordsByType(ctx context.Context, recordType string, limit, offset int) ([]*types.HistoryRecord, error) {
	var records []*types.HistoryRecord
	query := s.db.Where("type = ?", recordType).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get records by type: %w", err)
	}
	return records, nil
}

// GetRecordsByTimeRange 按时间范围获取历史记录
func (s *HistoryService) GetRecordsByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*types.HistoryRecord, error) {
	var records []*types.HistoryRecord
	query := s.db.Where("created_at BETWEEN ? AND ?", start, end).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get records by time range: %w", err)
	}
	return records, nil
}

// DeleteRecord 删除历史记录
func (s *HistoryService) DeleteRecord(ctx context.Context, id string) error {
	result := s.db.Delete(&types.HistoryRecord{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("record not found: %s", id)
	}
	return nil
}

// ClearRecords 清空历史记录
func (s *HistoryService) ClearRecords(ctx context.Context) error {
	return s.db.Where("1 = 1").Delete(&types.HistoryRecord{}).Error
}

// ClearOldRecords 清理旧记录（保留最近N天的记录）
func (s *HistoryService) ClearOldRecords(ctx context.Context, days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	return s.db.Where("created_at < ?", cutoff).Delete(&types.HistoryRecord{}).Error
}
