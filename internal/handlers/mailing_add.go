package handlers

import (
	"net/http"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AddSubRequest Структура запроса
type AddSubRequest struct {
	BirthDay	string	`json:"birth_day" binding:"required"`
	Email		string	`json:"email" binding:"required"`
	Firstname	string	`json:"firstname" binding:"required"`
	Lastname	string	`json:"lastname" binding:"required"`
}

// AddSub Добавить подписчика в рассылку
func (h *MailingHandler) AddSub(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "AddSubHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("AddSubHandler() started")
	defer l.Debug("AddSubHandler() done")

	var req AddSubRequest

	// Если данные не прошли валидацию, то просто выходим из "ручки", т.к. в bindData уже записана ошибка
	// через ctx.JSON...
	if ok := bindData(ctx, &req); !ok {
		return
	}

	// Валидируем почту
	if !ValidEmail(RegMail, req.Email) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid email",
		})

		return
	}

	// Парсим дату дня рождения
	BD, err := ParseDate(req.BirthDay)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid date format",
		})

		return
	}

	// Маппим данные в структуру DTO
	sub := models.Sub{
		BirthDay:	BD,
		Firstname:	req.Firstname,
		Lastname:	req.Lastname,
		Email:		req.Email,
	}

	// Добавляем подписчика
	if err = h.mailingService.AddSubTemplate(ctx, sub); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	ctx.JSON(http.StatusOK, req)
}
