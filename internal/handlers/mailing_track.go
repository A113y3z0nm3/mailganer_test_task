package handlers

import (
	"fmt"
	"net/http"

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
	// Получаем уникальный номер письма
	uuid, err := getUUIDFromParam(ctx)
	if err != nil {

	}

	uid := fmt.Sprint(uuid)

	// Записываем событие открытия
	err = h.trackService.WriteOpening(ctx, uid)
	if err != nil {

	}

	// Передаем изображение клиенту
	image, err := h.trackService.GetImage(ctx)
	if err != nil {

	}

	ctx.Data(http.StatusOK, "image", image)
}
