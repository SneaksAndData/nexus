package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetRunResultsByTag godoc
//
//	@Summary		Read run results by tag
//	@Description	Read results of all runs with a matching tag
//	@Tags			results
//	@Produce		json
//	@Produce		plain
//	@Param			requestTag	path		string	true	"Request tag assigned by a client"
//	@Success		200	{array}    models.TaggedRequestResult
//	@Failure		400	{string}	string
//	@Failure		404	{string}	string
//	@Router			/algorithm/v1.2/results/tags/{requestTag} [get]
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

		responseContent := []*models.TaggedRequestResult{}
		for result := range results {
			responseContent = append(responseContent, models.NewTaggedRequestResult(result))
		}

		ctx.JSON(http.StatusOK, responseContent)
	}
}
