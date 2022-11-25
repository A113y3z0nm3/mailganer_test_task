package services

import (
	"context"
	"mailganer_test_task/internal/models"
)

// MailingServiceConfig
type MailingServiceConfig struct {

}

// MailingService
type MailingService struct {

}

// NewMailingService
func NewMailingService(c *MailingServiceConfig) *MailingService {
	return &MailingService{

	}
}

// PushEmail
func (s *MailingService) PushEmail(context.Context) (models.Mailing, error) {

}
