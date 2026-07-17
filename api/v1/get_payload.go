package v1

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/urlsign"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// GetRunPayload godoc
//
//		@Summary		Read a run payload
//		@Description	Retrieves payload sent by the client for the provided run
//		@Tags			payload
//		@Produce		plain
//		@Produce		html
//	 	@Produce        octet-stream
//		@Param			algorithmName	path		string	true	"Algorithm name"
//		@Param			requestId	path		string	true	"Request identifier"
//		@Success		200	{object}    interface{}
//		@Failure		400	{string}	string
//		@Failure		403	{string}	string
//		@Failure		404	{string}	string
//		@Failure		401	{string}	string
//		@Router			/algorithm/v1/payload/{algorithmName}/requests/{requestId} [get]
func GetRunPayload(buffer request.Buffer, logger klog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")
		parsed, _ := url.Parse(ctx.Request.URL.String())

		// TODO: add secret here
		err := urlsign.Verify(*parsed, []byte{})

		if err != nil {
			logger.V(0).Error(err, "Unauthorized payload url: %s", parsed.String())
			ctx.String(http.StatusForbidden, `Invalid payload address: %s`, parsed.String())
			return
		}

		result, err := buffer.GetPersisted(requestId, algorithmName)

		if err != nil {
			logger.V(0).Error(err, "Failure when reading a persisted payload", "requestId", requestId, "algorithmName", algorithmName)
			ctx.String(http.StatusBadRequest, `Failed to find a run for %s`, requestId)
			return
		}

		if result == nil {
			ctx.String(http.StatusNotFound, "")
			return
		}

		var payloadObj interface{}
		err = json.Unmarshal(result, &payloadObj)

		if err != nil {
			logger.V(0).Error(err, "Failure when parsing a persisted payload", "requestId", requestId, "algorithmName", algorithmName)
			ctx.String(http.StatusBadRequest, `Failed to parse payload: %s`, result)
			return
		}

		ctx.JSON(http.StatusOK, payloadObj)
	}
}
