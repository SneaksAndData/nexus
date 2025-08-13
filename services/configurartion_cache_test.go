package services

import (
	"context"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	"testing"
	"time"
)

var resyncPeriod = time.Second * 0

type fixture struct {
	client      nexuscore.Interface
	configCache *NexusResourceCache
	ctx         context.Context
}

func newFixture(t *testing.T, existingObjects []runtime.Object) *fixture {
	f := &fixture{}
	_, ctx := ktesting.NewTestContext(t)

	f.client = fake.NewClientset(existingObjects...)
	f.ctx = ctx

	f.configCache = NewNexusResourceCache(f.client, "test", klog.FromContext(ctx), &resyncPeriod)
	return f
}

func (f *fixture) populateTemplates(templates []*v1.NexusAlgorithmTemplate) {
	for _, template := range templates {
		_ = f.configCache.templateInformer.GetIndexer().Add(template)
	}
}

func (f *fixture) populateWorkgroups(workgroups []*v1.NexusAlgorithmWorkgroup) {
	for _, workgroup := range workgroups {
		_ = f.configCache.workgroupInformer.GetIndexer().Add(workgroup)
	}
}

func TestNexusResourceCache_GetAlgorithmConfiguration(t *testing.T) {
	templates := []*v1.NexusAlgorithmTemplate{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NexusAlgorithmTemplate",
				APIVersion: "science.sneaksanddata.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-algorithm-1",
				Namespace: "test",
			},
			Spec:   v1.NexusAlgorithmSpec{},
			Status: v1.NexusAlgorithmStatus{},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NexusAlgorithmTemplate",
				APIVersion: "science.sneaksanddata.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-algorithm-2",
				Namespace: "test",
			},
			Spec:   v1.NexusAlgorithmSpec{},
			Status: v1.NexusAlgorithmStatus{},
		},
	}
	existingConfigurations := []runtime.Object{}

	for _, template := range templates {
		existingConfigurations = append(existingConfigurations, template)
	}

	f := newFixture(t, existingConfigurations)

	f.populateTemplates(templates)

	err := f.configCache.Init(f.ctx)

	if err != nil {
		t.Errorf("failed to init configuration cache: %v", err)
		t.FailNow()
	}

	config, err := f.configCache.GetAlgorithmConfiguration("test-algorithm-1")

	if err != nil {
		t.Errorf("failed to get algorithm configuration: %v", err)
		t.FailNow()
	}

	if config == nil {
		t.Errorf("algorithm configuration should not be nil")
	}
}

func TestNexusResourceCache_GetDeletedAlgorithmConfiguration(t *testing.T) {
	templates := []*v1.NexusAlgorithmTemplate{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NexusAlgorithmTemplate",
				APIVersion: "science.sneaksanddata.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-algorithm-1",
				Namespace: "test",
			},
			Spec:   v1.NexusAlgorithmSpec{},
			Status: v1.NexusAlgorithmStatus{},
		},
	}

	f := newFixture(t, []runtime.Object{})

	f.populateTemplates(templates)

	err := f.configCache.Init(f.ctx)

	if err != nil {
		t.Errorf("failed to init configuration cache: %v", err)
		t.FailNow()
	}

	config, err := f.configCache.GetAlgorithmConfiguration("test-algorithm-1")

	if err != nil {
		t.Errorf("failed to get algorithm configuration: %v", err)
		t.FailNow()
	}

	if config != nil {
		t.Errorf("algorithm configuration should be nil")
	}
}

func TestNexusResourceCache_GetWorkgroup(t *testing.T) {
	workgroups := []*v1.NexusAlgorithmWorkgroup{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NexusAlgorithmWorkgroup",
				APIVersion: "science.sneaksanddata.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workgroup-1",
				Namespace: "test",
			},
			Spec: v1.NexusAlgorithmWorkgroupSpec{},
		},
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NexusAlgorithmWorkgroup",
				APIVersion: "science.sneaksanddata.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-workgroup-2",
				Namespace: "test",
			},
			Spec: v1.NexusAlgorithmWorkgroupSpec{},
		},
	}

	existingConfigurations := []runtime.Object{}

	for _, workgroup := range workgroups {
		existingConfigurations = append(existingConfigurations, workgroup)
	}

	f := newFixture(t, existingConfigurations)

	f.populateWorkgroups(workgroups)

	err := f.configCache.Init(f.ctx)

	if err != nil {
		t.Errorf("failed to init configuration cache: %v", err)
		t.FailNow()
	}

	config, err := f.configCache.GetWorkgroupConfiguration("test-workgroup-2")

	if err != nil {
		t.Errorf("failed to get algorithm workgroup: %v", err)
		t.FailNow()
	}

	if config == nil {
		t.Errorf("algorithm workgroup should not be nil")
	}
}
