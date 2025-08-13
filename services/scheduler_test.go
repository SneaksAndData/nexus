package services

import (
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus/services/models"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"testing"
)

type schedulerFixture struct {
	scheduler  *RequestScheduler
	kubeClient kubernetes.Interface
	buffer     request.Buffer
}

func newSchedulerFixture(t *testing.T) *schedulerFixture {
	_, ctx := ktesting.NewTestContext(t)
	f := &schedulerFixture{}

	f.kubeClient = k8sfake.NewClientset()
	f.buffer = request.NewMemoryPassthroughBuffer(ctx, map[string]string{})

	f.scheduler = NewRequestScheduler(&models.PipelineWorkerConfig{
		FailureRateBaseDelay:       0,
		FailureRateMaxDelay:        0,
		RateLimitElementsPerSecond: 0,
		RateLimitElementsBurst:     0,
		Workers:                    0,
	}, f.kubeClient, []*shards.ShardClient{}, f.buffer, "nexus", klog.FromContext(ctx), &resyncPeriod)

	return f
}
