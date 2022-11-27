package handlers

import (
	"net/http"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getUIDFromParam Получает уникальный номер из ссылки
func getUUIDFromParam(ctx *gin.Context) (uuid.UUID, error) {
	str := ctx.Param("uid")

	uuid, err := uuid.Parse(str)

	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

// Track Отслеживает открытие сообщения и передает изображение в письмо
func (h *MailingHandler) Track(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "TrackHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("TrackHandler() started")
	defer l.Debug("TrackHandler() done")

	// Получаем уникальный номер письма
	uid, err := getUUIDFromParam(ctx)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

		return
	}

	// Записываем событие открытия
	h.mailingService.WriteOpening(ctx, uid)

	// Передаем изображение клиенту
	ctx.Header("Content-Type", "text/html")
	ctx.File("../images/mailing.jpeg")
}
