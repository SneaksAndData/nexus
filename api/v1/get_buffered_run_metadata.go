package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetBufferedRunMetadata godoc
//
//	@Summary		Read a buffered run metadata (Kubernetes Job JSON)
//	@Description	Retrieves a buffered metadata for a run
//	@Tags			metadata
//	@Produce		json
//	@Produce		plain
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId	path		string	true	"Request identifier"
//	@Success		200	{string}	string
//	@Failure		400	{string}	string
//	@Failure		404	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/buffer/{algorithmName}/requests/{requestId} [get]
func GetBufferedRunMetadata(buffer request.Buffer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")

		result, err := buffer.GetBufferedEntry(&models.CheckpointedRequest{
			Algorithm: algorithmName,
			Id:        requestId,
		})

		if err != nil {
			ctx.String(http.StatusBadRequest, `Failed to read buffered metadata for %s`, requestId)
			return
		}

		if result == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		ctx.String(http.StatusOK, result.Template)
	}
}
