package handlers

import (
	"net/http"
	"time"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

// SendMailRequest Структура запроса
type SendMailRequest struct {
	BirthDay	string	`json:"birth_day"`
	Email		string	`json:"email" binding:"required"`
	Firstname	string	`json:"firstname" binding:"required"`
	Lastname	string	`json:"lastname" binding:"required"`
}

// SendMail Отправить письмо
func (h *MailingHandler) SendMail(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "SendMailHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("SendMailHandler() started")
	defer l.Debug("SendMailHandler() done")

	var req SendMailRequest

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

	// Парсим дату дня рождения, если она указана
	var BD time.Time
	var err error
	if req.BirthDay != "" {
		BD, err = ParseDate(req.BirthDay)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid date format",
			})

			return
		}
	}

	// Маппим данные в DTO структуру
	sub := models.Sub{
		BirthDay:	BD,
		Email:		req.Email,
		Firstname:	req.Firstname,
		Lastname:	req.Lastname,
	}

	// Отправляем рассылочное письмо
	if err := h.mailingService.PushEmail(ctxLog, sub); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})

		return
	}

	ctx.JSON(http.StatusOK, "message sent")
}
