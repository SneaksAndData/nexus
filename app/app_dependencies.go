package app

import (
	"context"
	"github.com/SneaksAndData/nexus-core/pkg/checkpoint/request"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	"github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned/scheme"
	nexusscheme "github.com/Sne
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

type ApplicationServices struct {
	checkpointBuffer *request.DefaultBuffer
	kubeClient       *kubernetes.Clientset
	nexusClient      *nexuscore.Clientset
	recorder         record.EventRecorder
}

func (services *ApplicationServices) WithBuffer(ctx context.Context) *ApplicationServices {
	if services.checkpointBuffer == nil {
		services.checkpointBuffer = request.NewDefaultBuffer(ctx, nil)
	}

	return services
}

func (services *ApplicationServices) WithKubeClients(ctx context.Context) *ApplicationServices {
	if services.kubeClient == nil || services.nexusClient == nil {
		logger := klog.FromContext(ctx)
		kubeCfg, err := clientcmd.BuildConfigFromFlags("", "")
		if err != nil {
			logger.Error(err, "Error building in-cluster kubeconfig for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		services.kubeClient, err = kubernetes.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "Error building in-cluster kubernetes clientset for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}

		services.nexusClient, err = nexuscore.NewForConfig(kubeCfg)
		if err != nil {
			logger.Error(err, "Error building in-cluster Nexus clientset for the scheduler")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}

	return services
}

func (services *ApplicationServices) WithRecorder(ctx context.Context, resourceNamespace string) *ApplicationServices {
	if services.recorder == nil {
		logger := klog.FromContext(ctx)
		// Create event broadcaster
		// Add nexus-configuration-controller types to the default Kubernetes Scheme so Events can be
		// logged for nexus-configuration-controller types.
		utilruntime.Must(nexusscheme.AddToScheme(scheme.Scheme))
		logger.V(4).Info("Creating event broadcaster")

		eventBroadcaster := record.NewBroadcaster(record.WithContext(ctx))
		eventBroadcaster.StartStructuredLogging(0)
		eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: services.kubeClient.CoreV1().Events(resourceNamespace)})

		services.recorder = eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "nexus"})
	}

	return services
}

func (services *ApplicationServices) CheckpointBuffer() *request.DefaultBuffer {
	return services.checkpointBuffer
}

func (services *ApplicationServices) KubeClient() *kubernetes.Clientset {
	return services.kubeClient
}

func (services *ApplicationServices) NexusClient() *nexuscore.Clientset {
	return services.nexusClient
}
