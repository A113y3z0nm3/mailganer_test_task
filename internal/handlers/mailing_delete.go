package handlers

import (
	log "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

//
type MailingDeleteSubRequest struct {

}

//
func (h *MailingHandler) DeleteSub(ctx *gin.Context) {
	ctxLog := log.ContextWithSpan(ctx, "DeleteSubHandler")
	l := h.logger.WithContext(ctxLog)

	l.Debug("DeleteSubHandler() started")
	defer l.Debug("DeleteSubHandler() done")

	
}
