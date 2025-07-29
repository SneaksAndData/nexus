package app

import (
	"context"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusscheme "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/scheme"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus/services"
	"github.com/SneaksAndData/nexus/services/models"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

type ApplicationServices struct {
	checkpointBuffer *request.DefaultBuffer
	defaultNamespace string
	kubeClient       *kubernetes.Clientset
	shardClients     []*shards.ShardClient
	nexusClient      *nexuscore.Clientset
	recorder         record.EventRecorder
	configCache      *services.NexusResourceCache
	scheduler        *services.RequestScheduler
	workerConfig     *models.PipelineWorkerConfig
}

func (appServices *ApplicationServices) WithAstraS3Buffer(ctx context.Context, config *request.S3BufferConfig, bundleConfig *request.AstraBundleConfig) *ApplicationServices {
	if appServices.checkpointBuffer == nil {
		appServices.checkpointBuffer = request.NewAstraS3Buffer(ctx, config, bundleConfig, map[string]string{})
		appServices.workerConfig = models.FromBufferConfig(config.BufferConfig)
	}

	return appServices
}

func (appServices *ApplicationServices) WithScyllaS3Buffer(ctx context.Context, config *request.S3BufferConfig, scyllaConfig *request.ScyllaCqlStoreConfig) *ApplicationServices {
	if appServices.checkpointBuffer == nil {
		appServices.checkpointBuffer = request.NewScyllaS3Buffer(ctx, config, scyllaConfig, map[string]string{})
		appServices.workerConfig = models.FromBufferConfig(config.BufferConfig)
	}

	return appServices
}

func (appServices *ApplicationServices) WithKubeClients(ctx context.Context, kubeConfigPath string) *ApplicationServices {
	if appServices.kubeClient == nil || appServices.nexusClient == nil {
		logger := klog.FromContext(ctx)
		kubeCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			logger.Error(err, "error building in-cluster kubeconfig for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		appServices.kubeClient, err = kubernetes.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "error building in-cluster kubernetes clientset for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		appServices.nexusClient, err = nexuscore.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "error building in-cluster Nexus clientset for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	return appServices
}

func (appServices *ApplicationServices) WithShards(ctx context.Context, shardConfigPath string, namespace string) *ApplicationServices {
	if appServices.shardClients == nil {
		logger := klog.FromContext(ctx)
		var shardLoaderError error
		appServices.shardClients, shardLoaderError = shards.LoadClients(shardConfigPath, namespace, logger)
		if shardLoaderError != nil {
			logger.Error(shardLoaderError, "unable to initialize shard clients")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	return appServices
}

func (appServices *ApplicationServices) WithDefaultNamespace(namespace string) *ApplicationServices {
	appServices.defaultNamespace = namespace
	return appServices
}

func (appServices *ApplicationServices) WithRecorder(ctx context.Context, resourceNamespace string) *ApplicationServices {
	if appServices.recorder == nil {
		logger := klog.FromContext(ctx)
		// Create event broadcaster
		// Add nexus-configuration-controller types to the default Kubernetes Scheme so Events can be
		// logged for nexus-configuration-controller types.
		utilruntime.Must(nexusscheme.AddToScheme(scheme.Scheme))
		logger.V(4).Info("creating event broadcaster")

		eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
		eventBroadcaster.StartStructuredLogging(0)
		eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: appServices.kubeClient.CoreV1().Events(resourceNamespace)})

		appServices.recorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "nexus"})
	}

	return appServices
}

func (appServices *ApplicationServices) WithCache(ctx context.Context, resourceNamespace string) *ApplicationServices {
	if appServices.configCache == nil {
		logger := klog.FromContext(ctx)
		appServices.configCache = services.NewNexusResourceCache(appServices.nexusClient, resourceNamespace, logger)
	}

	return appServices
}

func (appServices *ApplicationServices) BuildScheduler(ctx context.Context) *ApplicationServices {
	logger := klog.FromContext(ctx)
	var err error

	appServices.scheduler, err = services.
		NewRequestScheduler(appServices.workerConfig, appServices.kubeClient, appServices.shardClients, appServices.checkpointBuffer, appServices.defaultNamespace, logger).
		Init(ctx)

	if err != nil {
		logger.Error(err, "unable to request scheduler")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return appServices
}

func (appServices *ApplicationServices) CheckpointBuffer() *request.DefaultBuffer {
	return appServices.checkpointBuffer
}

func (appServices *ApplicationServices) Logger(ctx context.Context) klog.Logger {
	return klog.FromContext(ctx)
}

func (appServices *ApplicationServices) KubeClient() *kubernetes.Clientset {
	return appServices.kubeClient
}

func (appServices *ApplicationServices) NexusClient() *nexuscore.Clientset {
	return appServices.nexusClient
}

func (appServices *ApplicationServices) Cache() *services.NexusResourceCache {
	return appServices.configCache
}

func (appServices *ApplicationServices) ShardClients() []*shards.ShardClient {
	return appServices.shardClients
}

func (appServices *ApplicationServices) Start(ctx context.Context) {
	logger := klog.FromContext(ctx)
	err := appServices.configCache.Init(ctx)
	if err != nil {
		logger.Error(err, "error building in-cluster kubeconfig for the scheduler")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	go appServices.scheduler.CommitActor.Start(ctx)
	go appServices.scheduler.SchedulerActor.Start(ctx)
	appServices.checkpointBuffer.Start(appServices.scheduler.SchedulerActor)
}
