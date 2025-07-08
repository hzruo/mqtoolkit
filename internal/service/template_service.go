package service

import (
	"context"
	"fmt"
	"mq-toolkit/pkg/types"
	"mq-toolkit/pkg/utils"

	"gorm.io/gorm"
)

// TemplateService manages message templates
type TemplateService struct {
	db *gorm.DB
}

// NewTemplateService creates a new TemplateService
func NewTemplateService(db *gorm.DB) *TemplateService {
	return &TemplateService{db: db}
}

// ListTemplates returns all message templates
func (s *TemplateService) ListTemplates(ctx context.Context) ([]*types.MessageTemplate, error) {
	var templates []*types.MessageTemplate
	if err := s.db.Order("created_at desc").Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to list templates: %w", err)
	}
	return templates, nil
}

// GetTemplate returns a single message template by ID
func (s *TemplateService) GetTemplate(ctx context.Context, id string) (*types.MessageTemplate, error) {
	var template types.MessageTemplate
	if err := s.db.First(&template, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}
	return &template, nil
}

// CreateTemplate creates a new message template
func (s *TemplateService) CreateTemplate(ctx context.Context, name, content string) (*types.MessageTemplate, error) {
	template := &types.MessageTemplate{
		ID:      utils.GenerateID(),
		Name:    name,
		Content: content,
	}
	if err := s.db.Create(template).Error; err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}
	return template, nil
}

// UpdateTemplate updates an existing message template
func (s *TemplateService) UpdateTemplate(ctx context.Context, id, name, content string) error {
	template := &types.MessageTemplate{
		ID:      id,
		Name:    name,
		Content: content,
	}
	return s.db.Model(&template).Updates(template).Error
}

// DeleteTemplate deletes a message template
func (s *TemplateService) DeleteTemplate(ctx context.Context, id string) error {
	return s.db.Delete(&types.MessageTemplate{}, "id = ?", id).Error
}