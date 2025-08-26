package models

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"time"
)

type PipelineWorkerConfig struct {
	FailureRateBaseDelay       time.Duration
	FailureRateMaxDelay        time.Duration
	RateLimitElementsPerSecond int
	RateLimitElementsBurst     int
	Workers                    int
}

func FromBufferConfig(bufferConfig *request.BufferConfig) *PipelineWorkerConfig { // coverage-ignore
	return &PipelineWorkerConfig{
		FailureRateBaseDelay:       bufferConfig.FailureRateBaseDelay,
		FailureRateMaxDelay:        bufferConfig.FailureRateMaxDelay,
		RateLimitElementsPerSecond: bufferConfig.RateLimitElementsPerSecond,
		RateLimitElementsBurst:     bufferConfig.RateLimitElementsBurst,
		Workers:                    bufferConfig.Workers,
	}
}
