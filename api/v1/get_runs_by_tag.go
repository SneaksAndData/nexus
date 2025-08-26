package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
	"net/http"
)

// GetRunResultsByTag godoc
//
//	@Summary		Read run results by tag
//	@Description	Read results of all runs with a matching tag
//	@Tags			results
//	@Produce		json
//	@Produce		plain
//	@Produce		html
//	@Param			requestTag	path		string	true	"Request tag assigned by a client"
//	@Success		200	{array}    models.TaggedRequestResult
//	@Failure		400	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/results/tags/{requestTag} [get]
func GetRunResultsByTag(buffer request.Buffer, logger klog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tag := ctx.Param("requestTag")

		results, err := buffer.GetTagged(tag)

		if err != nil {
			logger.V(0).Error(err, "Failed to read tagged results", "tag", tag)
			ctx.String(http.StatusBadRequest, `Failed to read tagged results for %s`, tag)
			return
		}

		responseContent := []*models.TaggedRequestResult{}
		for result := range results {
			responseContent = append(responseContent, models.NewTaggedRequestResult(result))
		}

		ctx.JSON(http.StatusOK, responseContent)
	}
}
