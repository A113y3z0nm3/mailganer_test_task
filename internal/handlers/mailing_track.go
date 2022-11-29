package handlers

import (
	"net/http"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

// Track Отслеживает открытие сообщения и передает изображение в письмо
func (h *MailingHandler) Track(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "TrackHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("TrackHandler() started")
	defer l.Debug("TrackHandler() done")

	// Получаем уникальный номер письма
	uid, err := getUUIDFromParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid UUID",
		})

		return
	}

	// Записываем событие открытия
	err = h.mailingService.WriteOpening(ctxLog, uid)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "sub not found",
		})

		return
	}

	// Передаем изображение клиенту
	ctx.Header("Content-Type", "text/html")
	ctx.File(h.imagePath)
}
