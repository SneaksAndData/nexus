package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRunPayload godoc
//
//	@Summary		Read a run payload
//	@Description	Retrieves payload sent by the client for the provided run
//	@Tags			payload
//	@Produce		plain
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId	path		string	true	"Request identifier"
//	@Success		302	{string}    string
//	@Failure		400	{string}	string
//	@Failure		404	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1.2/payload/{algorithmName}/requests/{requestId} [get]
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
