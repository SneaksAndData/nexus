package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

func CreateRun(buffer *request.DefaultBuffer, configCache *services.MachineLearningAlgorithmCache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// TODO: log errors
		algorithmName := ctx.Param("algorithmName")
		payload := models.AlgorithmRequest{}
		requestId := uuid.New()

		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.String(http.StatusBadRequest, `Algorithm payload is invalid: %s`, err.Error())
			return
		}

		config, cacheErr := configCache.GetConfiguration(algorithmName)

		if cacheErr != nil || config == nil {
			ctx.String(http.StatusBadRequest, `No valid configuration found for: %s. Please check that algorithm name is spelled correctly and try again. Contact an algorithm author if this problem persists.`, algorithmName)
		}

		if err := buffer.Add(requestId.String(), algorithmName, &payload, &config.Spec); err != nil {
			ctx.String(http.StatusBadRequest, `Request buffering failed for: %s, error: %s`, requestId.String(), err.Error())
			return
		}

		ctx.JSON(http.StatusAccepted, map[string]string{
			"requestId": requestId.String(),
		})

		return
	}
}
