package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

// MailingService Интерфейс к сервису, отвечающему за отслеживание писем
type MailingService interface {
	WriteOpening(context.Context, string) error
	GetImage(context.Context) ([]byte, error)
}

// MailingHandlerConfig Конфигурация для обработчика
type MailingHandlerConfig struct {
	Router			*gin.Engine
	TrackService	MailingService
}

// MailingHandler Обработчик для фиксации открытия письма и передачи в него изображения
type MailingHandler struct {
	trackService	MailingService
}

// RegisterMailingHandler Регистратор обработчика
func RegisterMailingHandler(c *MailingHandlerConfig) {
	mailingHandler := MailingHandler{
		trackService: c.TrackService,
	}

	g := c.Router.Group("v1") // Версия API

	
	g.GET("/:uid", mailingHandler.Track)			// Получить изображение и отследить получение
	g.POST("/newSub", mailingHandler.AddSub)		// Добавить подписчика в рассылку
	g.POST("/sendMail", mailingHandler.SendMail)	// Отправить рассылку списку подписчиков
	g.DELETE("/deleteSub", mailingHandler.Delete)	// Удалить подписчика из рассылки
}
