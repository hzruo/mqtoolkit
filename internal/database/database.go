package database

import (
	"fmt"
	"mq-toolkit/pkg/types"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库管理器
type Database struct {
	db *gorm.DB
}

// New 创建新的数据库实例
func New(dbPath string) (*Database, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{db: db}

	// 自动迁移数据库表
		if err := db.AutoMigrate(&types.ConnectionConfig{}, &types.HistoryRecord{}, &types.MessageTemplate{}); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	return database, nil
}

// migrate 执行数据库迁移
func (d *Database) migrate() error {
	return d.db.AutoMigrate(
		&types.ConnectionConfig{},
		&types.HistoryRecord{},
		&types.MessageTemplate{},
	)
}

// GetDB 获取数据库实例
func (d *Database) GetDB() *gorm.DB {
	return d.db
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

// Health 检查数据库健康状态
func (d *Database) Health() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Ping()
}

// Stats 获取数据库统计信息
func (d *Database) Stats() (map[string]interface{}, error) {
	sqlDB, err := d.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	stats := sqlDB.Stats()

	// 获取表记录数
	var connectionCount, historyCount, templateCount int64
	d.db.Model(&types.ConnectionConfig{}).Count(&connectionCount)
	d.db.Model(&types.HistoryRecord{}).Count(&historyCount)
	d.db.Model(&types.MessageTemplate{}).Count(&templateCount)

	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
		"connection_count":     connectionCount,
		"history_count":        historyCount,
		"template_count":       templateCount,
	}, nil
}

// Backup 备份数据库
func (d *Database) Backup(backupPath string) error {
	// 确保备份目录存在
	dir := filepath.Dir(backupPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// 执行 VACUUM INTO 命令进行备份
	if err := d.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath)).Error; err != nil {
		return fmt.Errorf("failed to backup database: %w", err)
	}

	return nil
}

// Restore 恢复数据库
func (d *Database) Restore(backupPath string) error {
	// 检查备份文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	// 关闭当前连接
	if err := d.Close(); err != nil {
		return fmt.Errorf("failed to close current database: %w", err)
	}

	// 这里需要在应用层面处理数据库文件的替换
	// 因为 GORM 不直接支持数据库恢复操作
	return fmt.Errorf("database restore should be handled at application level")
}

// Clean 清理数据库（删除所有数据但保留表结构）
func (d *Database) Clean() error {
	// 删除所有表的数据
	if err := d.db.Exec("DELETE FROM connection_configs").Error; err != nil {
		return fmt.Errorf("failed to clean connection_configs: %w", err)
	}

	if err := d.db.Exec("DELETE FROM history_records").Error; err != nil {
		return fmt.Errorf("failed to clean history_records: %w", err)
	}

	if err := d.db.Exec("DELETE FROM message_templates").Error; err != nil {
		return fmt.Errorf("failed to clean message_templates: %w", err)
	}

	// 重置自增ID
	if err := d.db.Exec("DELETE FROM sqlite_sequence").Error; err != nil {
		return fmt.Errorf("failed to reset auto increment: %w", err)
	}

	return nil
}

// Transaction 执行事务
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.db.Transaction(fn)
}
