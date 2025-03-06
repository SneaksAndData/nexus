package app

import (
	"context"
	"fmt"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/scheme"
	nexusscheme "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/scheme"
	"github.com/SneaksAndData/nexus-core/pkg/pipeline"
	"github.com/SneaksAndData/nexus-core/pkg/shards"
	"github.com/SneaksAndData/nexus/services"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"time"
)

type ApplicationServices struct {
	checkpointBuffer *request.DefaultBuffer
	defaultNamespace string
	kubeClient       *kubernetes.Clientset
	shardClients     []*shards.ShardClient
	nexusClient      *nexuscore.Clientset
	recorder         record.EventRecorder
	configCache      *services.MachineLearningAlgorithmCache
}

func (appServices *ApplicationServices) WithBuffer(ctx context.Context, config *request.BufferConfig, bundleConfig *request.AstraBundleConfig) *ApplicationServices {
	if appServices.checkpointBuffer == nil {
		appServices.checkpointBuffer = request.NewDefaultBuffer(ctx, config, bundleConfig, map[string]string{})
	}

	return appServices
}

func (appServices *ApplicationServices) WithKubeClients(ctx context.Context, kubeConfigPath string) *ApplicationServices {
	if appServices.kubeClient == nil || appServices.nexusClient == nil {
		logger := klog.FromContext(ctx)
		kubeCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			logger.Error(err, "Error building in-cluster kubeconfig for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		appServices.kubeClient, err = kubernetes.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "Error building in-cluster kubernetes clientset for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		appServices.nexusClient, err = nexuscore.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "Error building in-cluster Nexus clientset for the scheduler")
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
			logger.Error(shardLoaderError, "Unable to initialize shard clients")
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
		logger.V(4).Info("Creating event broadcaster")

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
		appServices.configCache = services.NewMachineLearningAlgorithmCache(appServices.nexusClient, resourceNamespace, logger)
	}

	return appServices
}

func (appServices *ApplicationServices) CheckpointBuffer() *request.DefaultBuffer {
	return appServices.checkpointBuffer
}

func (appServices *ApplicationServices) KubeClient() *kubernetes.Clientset {
	return appServices.kubeClient
}

func (appServices *ApplicationServices) NexusClient() *nexuscore.Clientset {
	return appServices.nexusClient
}

func (appServices *ApplicationServices) Cache() *services.MachineLearningAlgorithmCache {
	return appServices.configCache
}

func (appServices *ApplicationServices) ShardClients() []*shards.ShardClient {
	return appServices.shardClients
}

func (appServices *ApplicationServices) getShardByName(shardName string) *shards.ShardClient {
	for _, shard := range appServices.shardClients {
		if shard.Name == shardName {
			return shard
		}
	}

	return nil
}

func (appServices *ApplicationServices) schedule(output *request.BufferOutput) (types.UID, error) {
	if output == nil {
		return types.UID(""), fmt.Errorf("buffer is nil")
	}

	var job = output.Checkpoint.ToV1Job()
	var submitted *batchv1.Job
	var submitErr error

	if shard := appServices.getShardByName(output.Checkpoint.AppliedConfiguration.WorkgroupHost); shard != nil {
		submitted, submitErr = shard.SendJob(shard.Namespace, &job)
	} else {
		submitted, submitErr = appServices.kubeClient.BatchV1().Jobs(appServices.defaultNamespace).Create(context.TODO(), &job, v1.CreateOptions{})
	}

	if submitErr != nil {
		return "", submitErr
	}

	return submitted.UID, nil
}

func (appServices *ApplicationServices) Start(ctx context.Context) {
	logger := klog.FromContext(ctx)
	err := appServices.configCache.Init(ctx)
	if err != nil {
		logger.Error(err, "Error building in-cluster kubeconfig for the scheduler")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	submissionActor := pipeline.NewDefaultPipelineStageActor[*request.BufferOutput, types.UID](
		"checkpoint_buffer",
		map[string]string{},
		time.Second*1,
		time.Second*5,
		10,
		100,
		10,
		appServices.schedule,
		nil,
	)

	go submissionActor.Start(ctx)
	appServices.checkpointBuffer.Start(submissionActor)
}
