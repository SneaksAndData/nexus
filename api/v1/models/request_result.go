package models

import (
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

type RequestResult struct {
	RequestId       string `json:"requestId"`
	Status          string `json:"status"`
	ResultUri       string `json:"resultUri"`
	RunErrorMessage string `json:"runErrorMessage"`
}

// parseCheckpointError maps error to a standard error code
func parseCheckpointError(request *models.CheckpointedRequest) string {
	switch request.AlgorithmFailureCode {
	case models.CAJ011.ErrorName(), models.CAJ000.ErrorName(), models.CAJ012.ErrorName(), models.CB000.ErrorName():
		return fmt.Sprintf("%s: %s", request.AlgorithmFailureCode, request.AlgorithmFailureCause)
	default:
		return "Fatal error during execution."
	}
}

// FromCheckpointedRequest converts CheckpointedRequest to a simplified result model
func FromCheckpointedRequest(request *models.CheckpointedRequest) *RequestResult {
	if request == nil {
		return nil
	}
	switch request.LifecycleStage {
	case models.LifecyclestageCompleted:
		return &RequestResult{
			RequestId: request.Id,
			Status:    request.LifecycleStage,
			ResultUri: request.ResultUri,
		}
	case models.LifecyclestageFailed, models.LifecyclestageCancelled:
		return &RequestResult{
			RequestId:       request.Id,
			Status:          request.LifecycleStage,
			RunErrorMessage: parseCheckpointError(request),
		}
	default:
		return &RequestResult{
			RequestId: request.Id,
			Status:    request.LifecycleStage,
		}
	}
}
