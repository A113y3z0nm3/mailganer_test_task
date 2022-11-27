package handlers

import (
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

//
type MailingSendRequest struct {

}

//
func (h *MailingHandler) SendMail(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "SendMailHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("SendMailHandler() started")
	defer l.Debug("SendMailHandler() done")

	
}
