package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func CreateRun(buffer *request.DefaultBuffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		algorithmName := ctx.Param("algorithmName")
		payload := models.AlgorithmRequest{}
		requestId := uuid.New()

		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.String(http.StatusBadRequest, `Algorithm payload is invalid: %s`, err.Error())
		} else {
			if err := buffer.Add(requestId.String(), &payload, nil); err != nil {
				ctx.String(http.StatusBadRequest, `Request buffering failed for: %s, error: %s`, requestId.String(), err.Error())
			}
			// TODO: remove algorithm name from payload and init it similar to request id
			ctx.JSON(http.StatusAccepted, map[string]string{
				"requestId": requestId.String(),
			})
		}

		return
	}
}
