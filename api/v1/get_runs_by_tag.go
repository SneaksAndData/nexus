package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetRunResultsByTag(buffer *request.DefaultBuffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		tag := ctx.Param("requestTag")

		results, err := buffer.GetTagged(tag)

		if err != nil {
			ctx.String(http.StatusBadRequest, `Failed to read tagged results for %s`, tag)
			return
		}

		if results == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		responseContent := []*models.RequestResult{}
		for result := range results {
			responseContent = append(responseContent, models.FromCheckpointedRequest(result))
		}

		ctx.JSON(http.StatusOK, responseContent)
	}
}
