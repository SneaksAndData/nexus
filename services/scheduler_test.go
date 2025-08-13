package services

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	coremodels "github.com/SneaksAndData/nexus-core/pkg/checkpoint/models"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/fake"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus/services/models"
	"github.com/aws/smithy-go/ptr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"testing"
	"time"
)

type schedulerFixture struct {
	scheduler        *RequestScheduler
	kubeClient       kubernetes.Interface
	shardClient      kubernetes.Interface
	nexusShardClient nexuscore.Interface
	buffer           request.Buffer
	ctx              context.Context
}

func (f *schedulerFixture) populatePods(pods []corev1.Pod) {
	for _, pod := range pods {
		_ = f.scheduler.podInformer.GetIndexer().Add(&pod)
	}
}

func newFakeRequest() *coremodels.AlgorithmRequest {
	return &coremodels.AlgorithmRequest{
		AlgorithmParameters: map[string]interface{}{
			"parameterA": "a",
			"parameterB": "b",
		},
		CustomConfiguration: nil,
		RequestApiVersion:   "",
		Tag:                 "",
		ParentRequest:       nil,
		PayloadValidFor:     "24h",
	}
}

func newFakeSpec() *v1.NexusAlgorithmSpec {
	return &v1.NexusAlgorithmSpec{
		Container: &v1.NexusAlgorithmContainer{
			Image:              "test-image",
			Registry:           "test",
			VersionTag:         "v1.2.3",
			ServiceAccountName: "test-sa",
		},
		ComputeResources: &v1.NexusAlgorithmResources{
			CpuLimit:        "1000m",
			MemoryLimit:     "2000Mi",
			CustomResources: nil,
		},
		WorkgroupRef: &v1.NexusAlgorithmWorkgroupRef{
			Name:  "default",
			Group: "science.sneaksanddata.com/v1",
			Kind:  "NexusAlgorithmWorkgroup",
		},
		Command: "python",
		Args:    []string{"main.py", "--request-id=%s", "--sas-uri=%s"},
		RuntimeEnvironment: &v1.NexusAlgorithmRuntimeEnvironment{
			EnvironmentVariables: []corev1.EnvVar{
				{
					Name:  "TEST_ENV_VAR",
					Value: "TEST_VALUE",
				},
			},
			MappedEnvironmentVariables: []corev1.EnvFromSource{
				{
					SecretRef: &corev1.SecretEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-secret",
						},
					},
				},
			},
			Annotations:     nil,
			DeadlineSeconds: ptr.Int32(300),
			MaximumRetries:  ptr.Int32(10),
		},
		ErrorHandlingBehaviour:     nil,
		DatadogIntegrationSettings: &v1.NexusDatadogIntegrationSettings{MountDatadogSocket: ptr.Bool(true)},
	}
}

func newFakeWorkgroupSpec() *v1.NexusAlgorithmWorkgroupSpec {
	return &v1.NexusAlgorithmWorkgroupSpec{
		Description:  "test",
		Capabilities: nil,
		Cluster:      "test-shard",
		Tolerations:  nil,
		Affinity:     nil,
	}
}

func newSchedulerFixture(t *testing.T, existingObjects []runtime.Object) *schedulerFixture {
	_, ctx := ktesting.NewTestContext(t)
	f := &schedulerFixture{}
	f.ctx = ctx

	f.kubeClient = k8sfake.NewClientset(existingObjects...)
	f.shardClient = k8sfake.NewClientset()
	f.nexusShardClient = fake.NewClientset()
	f.buffer = request.NewMemoryPassthroughBuffer(ctx, map[string]string{})

	f.scheduler = NewRequestScheduler(&models.PipelineWorkerConfig{
		FailureRateBaseDelay:       time.Second,
		FailureRateMaxDelay:        time.Second * 2,
		RateLimitElementsPerSecond: 10,
		RateLimitElementsBurst:     10,
		Workers:                    2,
	}, f.kubeClient, []*shards.ShardClient{
		shards.NewShardClient(f.shardClient, f.nexusShardClient, "test-shard", "nexus"),
	}, f.buffer, "nexus", klog.FromContext(ctx), &resyncPeriod)

	return f
}

func TestScheduler(t *testing.T) {
	f := newSchedulerFixture(t, []runtime.Object{})
	scheduler, err := f.scheduler.Init(f.ctx)

	if err != nil {
		t.Errorf("scheduler init failed: %s", err)
		t.FailNow()
	}

	go f.scheduler.Start(f.ctx)
	go f.buffer.Start(scheduler.SchedulerActor)

	time.Sleep(1 * time.Second)

	err = f.buffer.Add("test", "test-algorithm", newFakeRequest(), newFakeSpec(), newFakeWorkgroupSpec())

	if err != nil {
		t.Errorf("failed to buffer an element: %s", err)
		t.FailNow()
	}

	// allow scheduling to happen
	time.Sleep(5 * time.Second)

	// check status
	checkpoint, err := f.buffer.Get("test", "test-algorithm")
	if err != nil {
		t.Errorf("failed to read a submitted run information: %s", err)
		t.FailNow()
	}

	if checkpoint == nil {
		t.Errorf("A checkpoint expected but none found")
		t.FailNow()
	}

	if checkpoint.LifecycleStage != coremodels.LifecycleStageRunning {
		t.Errorf("The checkpoint lifecycle stage must be running, but %s", checkpoint.LifecycleStage)
		t.FailNow()
	}

	// gracefully stop
	f.ctx.Done()
	time.Sleep(time.Second)
}

func TestScheduler_Restart(t *testing.T) {
	events := []corev1.Event{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Event",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-event",
				Namespace: "nexus",
			},
			InvolvedObject: corev1.ObjectReference{
				Name:      "test-terminated-scheduler",
				Namespace: "nexus",
				Kind:      "Pod",
			},
			Reason: "Evicted",
		},
	}
	terminatedPods := []corev1.Pod{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-terminated-scheduler",
				Namespace: "nexus",
				Labels: map[string]string{
					ComponentKey: ComponentName,
				},
			},
			Spec: corev1.PodSpec{},
			Status: corev1.PodStatus{
				Phase: corev1.PodSucceeded,
			},
		},
	}

	objects := []runtime.Object{}

	for _, event := range events {
		objects = append(objects, &event)
	}

	for _, pod := range terminatedPods {
		objects = append(objects, &pod)
	}

	f := newSchedulerFixture(t, objects)
	f.populatePods(terminatedPods)

	// add BUFFERED submission
	buffer := f.buffer.(*request.MemoryPassthroughBuffer)
	checkpoint, _, _ := coremodels.FromAlgorithmRequest("test-id", "test-algorithm", newFakeRequest(), newFakeSpec())
	checkpoint.LifecycleStage = coremodels.LifecycleStageBuffered
	checkpoint.ReceivedByHost = "test-terminated-scheduler"

	buffer.Checkpoints = append(buffer.Checkpoints, checkpoint)
	buffer.BufferedEntries = append(buffer.BufferedEntries, coremodels.FromCheckpoint(checkpoint, newFakeWorkgroupSpec()))

	_, err := f.scheduler.Init(f.ctx)

	if err != nil {
		t.Errorf("failed to initialize scheduler: %s", err)
		t.FailNow()
	}
	go f.scheduler.Start(f.ctx)

	// wait for the missed submission to be sent
	time.Sleep(5 * time.Second)

	// check status
	lateCheckpoint, err := f.buffer.Get("test-id", "test-algorithm")
	if err != nil {
		t.Errorf("failed to read a submitted run information: %s", err)
		t.FailNow()
	}

	if lateCheckpoint == nil {
		t.Errorf("A checkpoint expected but none found")
		t.FailNow()
	}

	if lateCheckpoint.LifecycleStage != coremodels.LifecycleStageRunning {
		t.Errorf("The checkpoint lifecycle stage must be running, but %s", lateCheckpoint.LifecycleStage)
		t.FailNow()
	}

	// gracefully stop
	f.ctx.Done()
	time.Sleep(time.Second)
}
