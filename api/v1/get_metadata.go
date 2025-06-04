package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRunMetadata godoc
//
//	@Summary		Read a run metadata
//	@Description	Retrieves checkpointed metadata for a run
//	@Tags			metadata
//	@Produce		json
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId	path		string	true	"Request identifier"
//	@Success		200	{object}	models.CheckpointedRequest
//	@Failure		400	{object}	string
//	@Failure		404	{object}	string
//	@Router			/metadata/{algorithmName}/requests/{requestId} [get]
func GetRunMetadata(buffer *request.DefaultBuffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")

		result, err := buffer.Get(requestId, algorithmName)

		if err != nil {
			ctx.String(http.StatusBadRequest, `Failed to read metadata for %s`, requestId)
			return
		}

		if result == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
