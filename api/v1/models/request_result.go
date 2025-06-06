package models

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
)

type RequestResult struct {
	RequestId       string `json:"requestId"`
	Status          string `json:"status"`
	ResultUri       string `json:"resultUri"`
	RunErrorMessage string `json:"runErrorMessage"`
}

type TaggedRequestResult struct {
	RequestId       string `json:"requestId"`
	AlgorithmName   string `json:"algorithmName"`
	Status          string `json:"status"`
	ResultUri       string `json:"resultUri"`
	RunErrorMessage string `json:"runErrorMessage"`
}

// FromCheckpointedRequest converts CheckpointedRequest to a simplified result model
func FromCheckpointedRequest(request *models.CheckpointedRequest) *RequestResult {
	if request == nil {
		return nil
	}
	switch request.LifecycleStage {
	case models.LifecycleStageCompleted:
		return &RequestResult{
			RequestId: request.Id,
			Status:    request.LifecycleStage,
			ResultUri: request.ResultUri,
		}
	case models.LifecycleStageFailed, models.LifecycleStageCancelled, models.LifecycleStageSchedulingFailed, models.LifecycleStageDeadlineExceeded:
		return &RequestResult{
			RequestId:       request.Id,
			Status:          request.LifecycleStage,
			RunErrorMessage: request.AlgorithmFailureCause,
		}
	default:
		return &RequestResult{
			RequestId: request.Id,
			Status:    request.LifecycleStage,
		}
	}
}

// NewTaggedRequestResult creates a new TaggedRequestResult from a CheckpointedRequest
func NewTaggedRequestResult(request *models.CheckpointedRequest) *TaggedRequestResult {
	if request == nil {
		return nil
	}
	return &TaggedRequestResult{
		AlgorithmName:   request.Algorithm,
		RequestId:       request.Id,
		Status:          request.LifecycleStage,
		ResultUri:       request.ResultUri,
		RunErrorMessage: request.AlgorithmFailureCause,
	}
}
