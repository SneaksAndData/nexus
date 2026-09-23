package v1

import (
	"net/http"

	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// UpdateRunTag godoc
//
//	@Summary		Assign a new client tag
//	@Description	Updates the specified run with a new client tag. Useful for performing a status reset on client side.
//	@Tags			metadata
//	@Produce		json
//	@Produce		plain
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId		path		string	true	"Request identifier"
//	@Param			newTag			body		string	true	"New client tag to assign"
//	@Success		200	{array}     string
//	@Failure		400	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/metadata/tags/{algorithmName}/requests/{requestId} [post]
func UpdateRunTag(buffer request.Buffer, logger klog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		algorithmName := ctx.Param("algorithmName")
		tagUpdateRequest := models.TagUpdateRequest{}
		requestId := ctx.Param("requestId")

		if err := ctx.ShouldBindJSON(&tagUpdateRequest); err != nil {
			ctx.String(http.StatusBadRequest, `Invalid tag update request: %s`, err.Error())
			return
		}

		checkpointToUpdate, err := buffer.Get(requestId, algorithmName)

		if err != nil {
			ctx.String(http.StatusInternalServerError, `Server error when updating tag`)
			logger.V(0).Error(err, "Failed to get checkpoint %s/%s to update its tag to %s", requestId, algorithmName, tagUpdateRequest.NewTag)
			return
		}

		if checkpointToUpdate == nil {
			ctx.String(http.StatusNotFound, "Checkpoint %s/%s not found", requestId, algorithmName)
			return
		}

		err = buffer.UpdateTag(checkpointToUpdate, tagUpdateRequest.NewTag)

		if err != nil {
			logger.V(0).Error(err, "Failed to update a tag", "tag", tagUpdateRequest.NewTag, "requestId", requestId, "algorithm", algorithmName)
			ctx.String(http.StatusBadRequest, `Failed to update a tag to requested value`)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{})
	}
}
