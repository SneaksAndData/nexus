package v1

import (
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	schedulermodels "github.com/SneaksAndData/nexus/api/v1/models"
	"github.com/SneaksAndData/nexus/services"
	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"net/http"
)

// CancelRun godoc
//
//	@Summary		Cancels an algorithm run
//	@Description	Interrupts the provided run id and cancels the execution tree if it exists
//	@Tags			cancellation
//	@Accept			json
//	@Produce		json
//	@Produce		plain
//	@Produce		html
//	@Param			algorithmName	path		string	true	"Algorithm name"
//	@Param			requestId	path		string	true	"Request identifier"
//	@Param			payload	body		schedulermodels.CancellationRequest	true	"Cancellation configuration"
//	@Success		200	{string}	string
//	@Failure		400	{string}	string
//	@Failure		500	{string}	string
//	@Failure		404	{string}	string
//	@Failure		401	{string}	string
//	@Router			/algorithm/v1/cancel/{algorithmName}/requests/{requestId} [post]
func CancelRun(buffer request.Buffer, scheduler *services.RequestScheduler, logger klog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		algorithmName := ctx.Param("algorithmName")
		requestId := ctx.Param("requestId")
		payload := schedulermodels.CancellationRequest{}

		if err := ctx.ShouldBindJSON(&payload); err != nil {
			ctx.String(http.StatusBadRequest, `Cancellation payload is invalid: %s`, err.Error())
			return
		}

		policy, err := payload.GetPolicy()

		if err != nil {
			ctx.String(http.StatusBadRequest, `Invalid cancellation request: %s`, err.Error())
			return
		}

		exists, err := scheduler.CancelRun(requestId, *policy)

		if !exists {
			ctx.String(http.StatusNotFound, `Provided request with identifier '%s' not found`, requestId)
			return
		}

		if !errors.IsNotFound(err) {
			ctx.String(http.StatusInternalServerError, `Unhandled error when executing a run cancellation. Please try again later`)
			logger.V(0).Error(err, "error when cancelling run %s/%s", algorithmName, requestId)
			return
		}

		// update status once the delete is done
		checkpoint, err := buffer.Get(requestId, algorithmName)
		if err != nil {
			ctx.String(http.StatusInternalServerError, `Error when reading a checkpoint for the cancelled request: %s/%s`, algorithmName, requestId)
			logger.V(0).Error(err, "checkpoint store failure when reading %s/%s", algorithmName, requestId)
			return
		}

		cancelled := checkpoint.DeepCopy()
		cancelled.LifecycleStage = models.LifecycleStageCancelled
		cancelled.AlgorithmFailureCause = fmt.Sprintf("Cancelled by '%s'", payload.Initiator)
		cancelled.AlgorithmFailureDetails = fmt.Sprintf("Run cancelled, reason: '%s'", payload.Reason)
		err = buffer.Update(cancelled)

		if err != nil {
			ctx.String(http.StatusInternalServerError, `Error when updating a checkpoint for the cancelled request: %s/%s`, algorithmName, requestId)
			logger.V(0).Error(err, "checkpoint store failure when updating %s/%s", algorithmName, requestId)
			return
		}

		ctx.String(http.StatusOK, "")
	}
}
