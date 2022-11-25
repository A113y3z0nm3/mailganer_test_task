package services

import (
	"context"
	"mailganer_test_task/internal/models"
)

// AdminServiceConfig
type AdminServiceConfig struct {

}

// AdminService
type AdminService struct {

}

// NewAdminService
func NewAdminService(c *AdminServiceConfig) *AdminService {
	return &AdminService{

	}
}

// GetAllSubscribers
func (s *AdminService) GetAllSubscribers(context.Context) ([]models.Sub, error) {

}

// GetSubscriber
func (s *AdminService) GetSubscriber(context.Context) (models.Sub, error) {
	
}

// AddSubscriber
func (s *AdminService) AddSubscriber(context.Context) (models.Sub, error) {
	
}

// EditSubscriber
func (s *AdminService) EditSubscriber(context.Context) (models.Sub, error) {
	
}

// RemoveSubscriber
func (s *AdminService) RemoveSubscriber(context.Context) (models.Sub, error) {
	
}
