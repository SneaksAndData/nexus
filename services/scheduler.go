package services

import (
	"context"
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/buildmeta"
	coremodels "github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/resolvers"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	"github.com/SneaksAndData/nexus/services/models"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"os"
	"time"
)

const (
	ComponentName = "scheduler"
	ComponentKey  = "app.kubernetes.io/component"
)

type LateSubmission struct {
	Checkpoint    *coremodels.CheckpointedRequest
	BufferedEntry *coremodels.SubmissionBufferEntry
}

type RequestScheduler struct {
	logger              klog.Logger
	workerConfig        *models.PipelineWorkerConfig
	factory             kubeinformers.SharedInformerFactory
	podInformer         cache.SharedIndexInformer
	eventInformer       cache.SharedIndexInformer
	LateSubmissionActor *pipeline.DefaultPipelineStageActor[*LateSubmission, *coremodels.CheckpointedRequest]
	SchedulerActor      *pipeline.DefaultPipelineStageActor[*request.BufferOutput, *coremodels.CheckpointedRequest]
	CommitActor         *pipeline.DefaultPipelineStageActor[*coremodels.CheckpointedRequest, string]
	shardClients        []*shards.ShardClient
	buffer              request.Buffer
}

func NewRequestScheduler(workerConfig *models.PipelineWorkerConfig, kubeClient kubernetes.Interface, shardClients []*shards.ShardClient, buffer request.Buffer, resourceNamespace string, logger klog.Logger, resyncPeriod *time.Duration) *RequestScheduler {
	defaultResyncPeriod := time.Second * 30
	factory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, *util.CoalescePointer(resyncPeriod, &defaultResyncPeriod), kubeinformers.WithNamespace(resourceNamespace))

	return &RequestScheduler{
		SchedulerActor: nil,
		workerConfig:   workerConfig,
		shardClients:   shardClients,
		factory:        factory,
		podInformer:    factory.Core().V1().Pods().Informer(),
		eventInformer:  factory.Core().V1().Events().Informer(),
		buffer:         buffer,
		logger:         logger,
	}
}

func (scheduler *RequestScheduler) Init(ctx context.Context) (*RequestScheduler, error) {
	scheduler.logger.Info("initializing Nexus scheduler")
	_, eventErr := scheduler.eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: scheduler.OnEvent,
	})

	if eventErr != nil {
		return nil, eventErr
	}

	scheduler.logger.Info("pod and event informers synced")

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

	scheduler.LateSubmissionActor = pipeline.NewDefaultPipelineStageActor[*LateSubmission, *coremodels.CheckpointedRequest](
		"late_submission",
		map[string]string{},
		scheduler.workerConfig.FailureRateBaseDelay,
		scheduler.workerConfig.FailureRateMaxDelay,
		scheduler.workerConfig.RateLimitElementsPerSecond,
		scheduler.workerConfig.RateLimitElementsBurst,
		scheduler.workerConfig.Workers,
		scheduler.lateSchedule,
		scheduler.CommitActor,
	)

	scheduler.logger.Info("actors configured")

	return scheduler, nil
}

func (scheduler *RequestScheduler) Start(ctx context.Context) {
	go scheduler.CommitActor.Start(ctx, nil)
	go scheduler.SchedulerActor.Start(ctx, nil)
	go scheduler.LateSubmissionActor.Start(ctx, pipeline.NewActorPostStart(func(ctx context.Context) error {
		scheduler.factory.Start(ctx.Done())

		if ok := cache.WaitForCacheSync(ctx.Done(), scheduler.podInformer.HasSynced, scheduler.eventInformer.HasSynced); !ok {
			return fmt.Errorf("failed to wait for pod self-informer caches to sync")
		}

		return nil
	}))
}

func (scheduler *RequestScheduler) OnEvent(obj interface{}) {
	if _, err := cache.ObjectToName(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	event := obj.(*corev1.Event)

	// ignore all events other than pods
	if event.InvolvedObject.Kind != "Pod" {
		return
	}

	pod, err := resolvers.GetCachedObject[corev1.Pod](event.InvolvedObject.Name, event.InvolvedObject.Namespace, scheduler.podInformer)

	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	// skip pods that no longer exist in informer cache
	if pod == nil {
		return
	}

	// check if pod is a Nexus scheduler instance
	if pod.Labels[ComponentKey] != ComponentName {
		return
	}

	// check that not self
	if podName, _ := os.Hostname(); podName == event.InvolvedObject.Name {
		return
	}

	// check that not active or pending
	if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodPending {
		return
	}

	// log and get reason
	scheduler.logger.V(0).Info("discovered an inactive scheduler instance", "instance", pod.Name)

	switch event.Reason {
	case "Killing", "Failed", "Terminated", "Evicted":
		scheduler.logger.V(0).Info("discovered a scheduler terminated externally", "instance", pod.Name, "reason", event.Reason, "message", event.Message)
		// host has been terminated - find its submissions and resubmit them
		checkpoints, err := scheduler.buffer.GetBuffered(event.InvolvedObject.Name)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		for checkpoint := range checkpoints {
			entry, err := scheduler.buffer.GetBufferedEntry(checkpoint)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				scheduler.LateSubmissionActor.Receive(&LateSubmission{
					Checkpoint:    checkpoint,
					BufferedEntry: entry,
				})
			}
		}
	default:
		scheduler.logger.V(0).Info("unmapped reason - skipping", "instance", pod.Name, "reason", event.Reason, "message", event.Message)
		return
	}
}

func (scheduler *RequestScheduler) commit(output *coremodels.CheckpointedRequest) (string, error) {
	output.LifecycleStage = coremodels.LifecycleStageRunning
	output.SentAt = time.Now()
	err := scheduler.buffer.Update(output)

	if err != nil {
		return output.Id, err
	}

	return output.Id, nil
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

func (scheduler *RequestScheduler) lateSchedule(submission *LateSubmission) (*coremodels.CheckpointedRequest, error) {
	if submission == nil {
		return nil, fmt.Errorf("no buffer entry provided")
	}

	job, err := submission.BufferedEntry.SubmissionTemplate()

	if err != nil {
		return nil, err
	}

	var submitted *batchv1.Job
	var submitErr error

	if shard := scheduler.getShardByName(submission.BufferedEntry.Cluster); shard != nil {
		submitted, submitErr = shard.SendJob(shard.Namespace, job)
	} else {
		return nil, fmt.Errorf("shard API server %s not configured", submission.BufferedEntry.Cluster)
	}

	if submitErr != nil {
		return nil, submitErr
	}

	resultCheckpoint := submission.Checkpoint.DeepCopy()
	resultCheckpoint.JobUid = string(submitted.UID)

	return resultCheckpoint, nil
}
