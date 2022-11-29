package handlers

import (
	"fmt"
	"net/http"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

// DeleteSub Удалить подписчика из рассылки
func (h *MailingHandler) DeleteSub(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "DeleteSubHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("DeleteSubHandler() started")
	defer l.Debug("DeleteSubHandler() done")

	// Проверяем и парсим UUID
	uid, err := getUUIDFromParam(ctx)

	// Если при парсинге UUID ошибка - закрываем
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid UUID",
		})

		return
	}

	sub := models.Sub{}

	sub.UUID = uid

	// Удаляем подписчика
	if err := h.mailingService.RemoveSub(ctxLog, sub); err != nil {
		if err.Error() == "sub not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("sub number %s has been deleted", uid.String()))
}
