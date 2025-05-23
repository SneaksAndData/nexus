package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRunPayload(buffer *request.DefaultBuffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")

		result, err := buffer.Get(requestId, algorithmName)

		if err != nil {
			ctx.String(http.StatusBadRequest, `Failed to find a run for %s`, requestId)
			return
		}

		if result == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		if result.PayloadUri == "" {
			ctx.String(http.StatusExpectationFailed, `Specified request %s does not have a serialized payload`, requestId)
			return
		}

		ctx.Redirect(http.StatusFound, result.PayloadUri)
	}
}
