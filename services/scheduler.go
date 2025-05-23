package services

import (
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/buildmeta"
	coremodels "github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus/services/models"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/klog/v2"
)

type RequestScheduler struct {
	logger         klog.Logger
	workerConfig   *models.PipelineWorkerConfig
	SchedulerActor *pipeline.DefaultPipelineStageActor[*request.BufferOutput, *coremodels.CheckpointedRequest]
	CommitActor    *pipeline.DefaultPipelineStageActor[*coremodels.CheckpointedRequest, string]
	shardClients   []*shards.ShardClient
	buffer         *request.DefaultBuffer
}

func NewRequestScheduler(workerConfig *models.PipelineWorkerConfig, shardClients []*shards.ShardClient, buffer *request.DefaultBuffer, logger klog.Logger) *RequestScheduler {
	return &RequestScheduler{
		SchedulerActor: nil,
		workerConfig:   workerConfig,
		shardClients:   shardClients,
		buffer:         buffer,
		logger:         logger,
	}
}

func (scheduler *RequestScheduler) Init() *RequestScheduler {
	scheduler.CommitActor = pipeline.NewDefaultPipelineStageActor[*coremodels.CheckpointedRequest, string](
		"commit",
		map[string]string{},
		scheduler.workerConfig.FailureRateBaseDelay,
		scheduler.workerConfig.FailureRateMaxDelay,
		scheduler.workerConfig.RateLimitElementsPerSecond,
		scheduler.workerConfig.RateLimitElementsBurst,
		scheduler.workerConfig.Workers,
		scheduler.commit,
		nil,
	)

	scheduler.SchedulerActor = pipeline.NewDefaultPipelineStageActor[*request.BufferOutput, *coremodels.CheckpointedRequest](
		"scheduler",
		map[string]string{},
		scheduler.workerConfig.FailureRateBaseDelay,
		scheduler.workerConfig.FailureRateMaxDelay,
		scheduler.workerConfig.RateLimitElementsPerSecond,
		scheduler.workerConfig.RateLimitElementsBurst,
		scheduler.workerConfig.Workers,
		scheduler.schedule,
		scheduler.CommitActor,
	)

	return scheduler
}

func (scheduler *RequestScheduler) commit(output *coremodels.CheckpointedRequest) (string, error) {
	toCommit := output.DeepCopy()
	toCommit.LifecycleStage = coremodels.LifecycleStageRunning
	err := scheduler.buffer.Update(toCommit)

	if err != nil {
		return toCommit.Id, err
	}

	return toCommit.Id, nil
}

func (scheduler *RequestScheduler) getShardByName(shardName string) *shards.ShardClient {
	for _, shard := range scheduler.shardClients {
		if shard.Name == shardName {
			return shard
		}
	}

	return nil
}

func (scheduler *RequestScheduler) schedule(output *request.BufferOutput) (*coremodels.CheckpointedRequest, error) {
	if output == nil {
		return nil, fmt.Errorf("buffer has not provided any data to schedule")
	}

	var job = output.Checkpoint.ToV1Job(fmt.Sprintf("%s-%s", buildmeta.AppVersion, buildmeta.BuildNumber), output.Workgroup)
	var submitted *batchv1.Job
	var submitErr error

	if shard := scheduler.getShardByName(output.Workgroup.Cluster); shard != nil {
		submitted, submitErr = shard.SendJob(shard.Namespace, &job)
	} else {
		return nil, fmt.Errorf("shard API server %s not configured", output.Workgroup.Cluster)
	}

	if submitErr != nil {
		return nil, submitErr
	}

	resultCheckpoint := output.Checkpoint.DeepCopy()
	resultCheckpoint.JobUid = string(submitted.UID)

	return resultCheckpoint, nil
}
