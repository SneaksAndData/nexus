package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRunResult(buffer *request.DefaultBuffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")

		result, err := buffer.Get(requestId, algorithmName)

		if err != nil {
			ctx.String(http.StatusBadRequest, `Failed to read results for %s`, requestId)
			return
		}

		if result == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		ctx.JSON(http.StatusOK, models.FromCheckpointedRequest(result))
	}
}
