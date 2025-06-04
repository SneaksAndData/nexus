package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

// CreateRun godoc
//
//	@Summary		Create a new algorithm run
//	@Description	Accepts an algorithm payload and places it into a scheduling queue
//	@Tags			run
//	@Accept			json
//	@Produce		json
//	@Produce		plain
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			payload	body		models.AlgorithmRequest	true	"Run configuration"
//	@Success		202	{object}	map[string]string
//	@Failure		400	{string}	string
//	@Failure		500	{string}	string
//	@Router			/algorithm/v1.2/run/{algorithmName} [post]
func CreateRun(buffer *request.DefaultBuffer, configCache *services.NexusResourceCache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		algorithmName := ctx.Param("algorithmName")
		payload := models.AlgorithmRequest{}
		requestId := uuid.New()

		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.String(http.StatusBadRequest, `Algorithm payload is invalid: %s`, err.Error())
			return
		}

		config, cacheErr := configCache.GetAlgorithmConfiguration(algorithmName)

		if cacheErr != nil {
			ctx.String(http.StatusInternalServerError, `Internal error occurred when processing your request.`, algorithmName)
			return
		}

		if config == nil {
			ctx.String(http.StatusBadRequest, `No valid configuration found for: %s. Please check that algorithm name is spelled correctly and try again. Contact an algorithm author if this problem persists.`, algorithmName)
			return
		}

		workgroup, err := configCache.GetWorkgroupConfiguration(config.Spec.WorkgroupRef.Name)

		if err != nil {
			ctx.String(http.StatusInternalServerError, `Internal error occurred when processing your request.`, algorithmName)
			return
		}

		if workgroup == nil {
			ctx.String(http.StatusBadRequest, `Cannot assign requested workgroup %s to the algorithm %s. Please check the deployed configuration.`, config.Spec.WorkgroupRef.Name, algorithmName)
			return
		}

		if err := buffer.Add(requestId.String(), algorithmName, &payload, &config.Spec, &workgroup.Spec); err != nil {
			ctx.String(http.StatusBadRequest, `Request buffering failed for: %s, error: %s`, requestId.String(), err.Error())
			return
		}

		ctx.JSON(http.StatusAccepted, map[string]string{
			"requestId": requestId.String(),
		})
	}
}
