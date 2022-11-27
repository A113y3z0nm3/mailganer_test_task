package handlers

import (
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

//
type MailingAddRequest struct {

}

//
func (h *MailingHandler) AddSub(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "AddSubHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("AddSubHandler() started")
	defer l.Debug("AddSubHandler() done")

	
}
