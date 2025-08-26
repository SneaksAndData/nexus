package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRunResult godoc
//
//	@Summary		Read a run result
//	@Description	Retrieves a result for the provided run
//	@Tags			results
//	@Produce		json
//	@Produce		plain
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId	path		string	true	"Request identifier"
//	@Success		200	{object}    models.RequestResult
//	@Failure		400	{object}	string
//	@Failure		404	{object}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/results/{algorithmName}/requests/{requestId} [get]
func GetRunResult(buffer request.Buffer) gin.HandlerFunc {
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
