package services

import (
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusinf "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	"k8s.io/client-go/tools/cache"
	"time"
)

type MachineLearningAlgorithmCache struct {
	configCache cache.ObjectName
}

func NewMachineLearningAlgorithmCache(client *nexuscore.Clientset, resourceNamespace string) *MachineLearningAlgorithmCache {
	factory := nexusinf.NewSharedInformerFactoryWithOptions(client, time.Second*30, nexusinf.WithNamespace(resourceNamespace))
	watcher := factory.Science().V1().MachineLearningAlgorithms()
	lister := watcher.Lister()
	hasSynced := watcher.Informer().HasSynced
}
