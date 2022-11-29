package handlers

import (
	"context"
	"regexp"
	"time"
	"mailganer_test_task/internal/models"
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MailingService Интерфейс к сервису, отвечающему за отслеживание писем
type MailingService interface {
	WriteOpening(ctx context.Context, uid uuid.UUID) error
	AddSubTemplate(ctx context.Context, sub models.Sub) error
	RemoveSub(ctx context.Context, sub models.Sub) error
	PushEmail(ctx context.Context, sub models.Sub) error
}

// MailingHandlerConfig Конфигурация для обработчика
type MailingHandlerConfig struct {
	Router			*gin.Engine
	MailingService	MailingService
	Logger			*log.Log
	ImagePath		string			`env:"HANDLER_IMAGE_PATH"`
}

// MailingHandler Обработчик для фиксации открытия письма и передачи в него изображения
type MailingHandler struct {
	mailingService	MailingService
	logger			*log.Log
	imagePath		string
}

// RegMail Шаблон проверки email-адреса
var RegMail = regexp.MustCompile(`^.+[@].+[.]\w+$`)

// ValidEmail Для валидации email в запросе
func ValidEmail(reg *regexp.Regexp, email string) bool {
	return reg.Match([]byte(email))
}

// ParseDate Для парсинга даты из запроса
func ParseDate(str string) (time.Time, error){
	return time.Parse("2006-01-01", str)
}

// getUIDFromParam Получает уникальный номер из ссылки
func getUUIDFromParam(ctx *gin.Context) (uuid.UUID, error) {
	str := ctx.Param("uid")

	uuid, err := uuid.Parse(str)

	if err != nil {
		return uuid, err
	}

	return uuid, nil
}

// RegisterMailingHandler Регистратор обработчика
func RegisterMailingHandler(c *MailingHandlerConfig) {
	mailingHandler := MailingHandler{
		mailingService: c.MailingService,
		logger:			c.Logger,
		imagePath:		c.ImagePath,
	}

	g := c.Router.Group("v1") // Версия API
	
	g.GET("/:uid", mailingHandler.Track)				// Получить изображение и отследить получение
	g.POST("/newSub", mailingHandler.AddSub)			// Добавить подписчика в рассылку
	g.POST("/sendMail", mailingHandler.SendMail)		// Отправить рассылку списку подписчиков
	g.DELETE("/:uid", mailingHandler.DeleteSub)	// Удалить подписчика из рассылки
}
