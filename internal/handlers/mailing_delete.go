package handlers

import (
	"net/http"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeleteSubRequest Структура запроса
type DeleteSubRequest struct {
	Email		string	`json:"email"`
	Uid			string	`json:"uid"`
}

// DeleteSub Удалить подписчика из рассылки
func (h *MailingHandler) DeleteSub(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "DeleteSubHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("DeleteSubHandler() started")
	defer l.Debug("DeleteSubHandler() done")

	var req DeleteSubRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Если данные о почте и UUID подписчика не указаны, отдаем ошибку, так как поиск возможен только по ним
	if (req.Email == "") && (req.Uid == "") {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "data is empty",
		})

		return
	}

	sub := models.Sub{}

	// Валидируем почту (если в запросе указана)
	if req.Email != "" {
		// Если почта не валидна и UUID отсутствует - отдаем ошибку
		if !ValidEmail(RegMail, req.Email) && (req.Uid == "") {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid email",
			})
	
			return
		}

		sub.Email = req.Email
	}

	// Проверяем и парсим UUID
	if req.Uid != "" {
		UUID, err := uuid.Parse(req.Uid)
		// Если при парсинге UUID ошибка, а email не валиден или отсутствует - отдаем ошибку
		if (err != nil) && ((req.Email == "") || !ValidEmail(RegMail, req.Email)) {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid data",
			})
	
			return
		}

		sub.UUID = UUID
	}

	// Удаляем подписчика
	if err := h.mailingService.RemoveSub(ctx, sub); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, req)
}
