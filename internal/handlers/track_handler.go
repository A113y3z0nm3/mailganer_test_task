package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

// TrackService Интерфейс к сервису, отвечающему за отслеживание писем
type TrackService interface {
	WriteOpening(context.Context, string) error
	GetImage(context.Context) ([]byte, error)
}

// TrackHandlerConfig Конфигурация для обработчика
type TrackHandlerConfig struct {
	Router			*gin.Engine
	TrackService	TrackService
}

// TrackHandler Обработчик для фиксации открытия письма и передачи в него изображения
type TrackHandler struct {
	trackService	TrackService
}

// RegisterTrackHandler Регистратор обработчика
func RegisterTrackHandler(c *TrackHandlerConfig) {
	trackHandler := TrackHandler{
		trackService: c.TrackService,
	}

	g := c.Router.Group("v1") // Версия API

	// Получить изображение и отследить получение
	g.GET("/:uid", trackHandler.Track)
}
