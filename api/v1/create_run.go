package v1

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			payload	body	models.AlgorithmRequest	true	"Run configuration"
//	@Param			dryRun	query	string	false	"If false, will buffer but not submit to the target cluster"
//	@Success		202	{object}	map[string]string
//	@Failure		400	{string}	string
//	@Failure		500	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/run/{algorithmName} [post]
func CreateRun(buffer request.Buffer, configCache *services.NexusResourceCache, scheduler *services.RequestScheduler, logger klog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var parentRef *metav1.OwnerReference

		algorithmName := ctx.Param("algorithmName")
		payload := models.AlgorithmRequest{}
		dryRun := ctx.DefaultQuery("dryRun", "false") == "true"
		requestId := uuid.New()

		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.String(http.StatusBadRequest, `Algorithm payload is invalid: %s`, err.Error())
			return
		}

		config, cacheErr := configCache.GetAlgorithmConfiguration(algorithmName)

		if cacheErr != nil {
			ctx.String(http.StatusInternalServerError, `Internal error occurred when processing your request.`, algorithmName)
			logger.V(0).Error(cacheErr, "error when retrieving algorithm template for %s/%s", algorithmName, requestId)
			return
		}

		if config == nil {
			ctx.String(http.StatusBadRequest, `No valid configuration found for: %s. Please check that algorithm name is spelled correctly and try again. Contact an algorithm author if this problem persists.`, algorithmName)
			return
		}

		workgroup, err := configCache.GetWorkgroupConfiguration(config.Spec.WorkgroupRef.Name)

		if err != nil {
			ctx.String(http.StatusInternalServerError, `Internal error occurred when processing your request.`, algorithmName)
			logger.V(0).Error(err, "error when retrieving algorithm workgroup configuration for %s/%s", algorithmName, requestId)
			return
		}

		if workgroup == nil {
			ctx.String(http.StatusBadRequest, `Cannot assign requested workgroup %s to the algorithm %s. Please check the deployed configuration.`, config.Spec.WorkgroupRef.Name, algorithmName)
			return
		}

		if payload.ParentRequest != nil {
			if !dryRun {
				parentRef, err = scheduler.ResolveParent(payload.ParentRequest.RequestId, workgroup.Spec.Cluster)

				if err != nil {
					ctx.String(http.StatusInternalServerError, `Internal error occurred when processing your request.`, algorithmName)
					logger.V(0).Error(err, "error when retrieving a parent request for %s/%s", algorithmName, requestId)
					return
				}
			} else {
				// for dry runs parent job might not exist at all, thus we create a fake reference
				parentRef = &metav1.OwnerReference{
					APIVersion: "batch/v1",
					Kind:       "Job",
					Name:       payload.ParentRequest.RequestId,
					UID:        types.UID(payload.ParentRequest.RequestId),
				}
			}

		}

		if err := buffer.Add(requestId.String(), algorithmName, &payload, &config.Spec, &workgroup.Spec, parentRef, dryRun); err != nil {
			ctx.String(http.StatusBadRequest, `Request buffering failed for: %s/%s`, algorithmName, requestId)
			logger.V(0).Error(err, "error when retrieving a parent request for %s/%s", algorithmName, requestId)
			return
		}

		ctx.JSON(http.StatusAccepted, map[string]string{
			"requestId": requestId.String(),
		})
	}
}
