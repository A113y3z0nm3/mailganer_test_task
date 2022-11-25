package handlers

import (
	"context"
	"mailganer_test_task/internal/models"

	"github.com/gin-gonic/gin"
)

// MailingService Интерфейс к сервису рассылок
type MailingService interface {
	PushEmail(context.Context) (models.Mailing, error)
}

// AdminService Интерфейс к сервису администрирования пользователей и рассылки
type AdminService interface {
	GetAllSubscribers(context.Context) ([]models.Sub, error)
	GetSubscriber(context.Context) (models.Sub, error)
	AddSubscriber(context.Context) (models.Sub, error)
	EditSubscriber(context.Context) (models.Sub, error)
	RemoveSubscriber(context.Context) (models.Sub, error)
}

// AdminHandlerConfig Конфигурация для AdminHandler
type AdminHandlerConfig struct {
	Router			*gin.Engine
	AdminService	AdminService
	MailingService	MailingService
}

// AdminHandler Обработчик запросов администрирования
type AdminHandler struct {
	adminService	AdminService
	mailingService	MailingService
}

// RegisterAdminHandler Регистратор обработчика
func RegisterAdminHandler(c *AdminHandlerConfig) {
	adminHandler := AdminHandler{
		adminService:	c.AdminService,
		mailingService: c.MailingService,
	}

	g := c.Router.Group("v1") // Версия API

	// Отправить рассылку
	g.POST("/push", adminHandler.Push)
	// Получить инфо обо всех подписчиках
	g.GET("/subs", adminHandler.GetAllSubs)
	// Получить инфо о конкретном подписчике
	g.GET("/subs/:sub", adminHandler.GetSub)
	// Добавить подписчика
	g.POST("/sub", adminHandler.AddSub)
	// Изменить подписчика
	g.PUT("/sub", adminHandler.EditSub)
	// Удалить подписчика
	g.DELETE("/sub", adminHandler.DeleteSub)
}
