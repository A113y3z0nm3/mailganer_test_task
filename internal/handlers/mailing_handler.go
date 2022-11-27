package handlers

import (
	"context"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MailingService Интерфейс к сервису, отвечающему за отслеживание писем
type MailingService interface {
	WriteOpening(ctx context.Context, uid uuid.UUID)
	AddSubTemplate(ctx context.Context, sub models.Sub) error
	RemoveSub(ctx context.Context, uid uuid.UUID, sub models.Sub) error
	PushEmail(ctx context.Context, sub models.Sub) error
}

// MailingHandlerConfig Конфигурация для обработчика
type MailingHandlerConfig struct {
	Router			*gin.Engine
	MailingService	MailingService
	Logger			*log.Log
}

// MailingHandler Обработчик для фиксации открытия письма и передачи в него изображения
type MailingHandler struct {
	mailingService	MailingService
	logger			*log.Log
}

// RegisterMailingHandler Регистратор обработчика
func RegisterMailingHandler(c *MailingHandlerConfig) {
	mailingHandler := MailingHandler{
		mailingService: c.MailingService,
	}

	g := c.Router.Group("v1") // Версия API

	
	g.GET("/:uid", mailingHandler.Track)			// Получить изображение и отследить получение
	g.POST("/newSub", mailingHandler.AddSub)		// Добавить подписчика в рассылку
	g.POST("/sendMail", mailingHandler.SendMail)	// Отправить рассылку списку подписчиков
	g.DELETE("/deleteSub", mailingHandler.DeleteSub)	// Удалить подписчика из рассылки
}
