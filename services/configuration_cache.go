package services

import (
	"context"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusinf "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	nexusinformers "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions/science/v1"
	nexuslisters "github.com/SneaksAndData/nexus-core/pkg/generated/listers/science/v1"
	"k8s.io/apimachinery/pkg/util/diff"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type MachineLearningAlgorithmCache struct {
	logger    klog.Logger
	factory   nexusinf.SharedInformerFactory
	watcher   nexusinformers.MachineLearningAlgorithmInformer
	lister    nexuslisters.MachineLearningAlgorithmLister
	hasSynced cache.InformerSynced
	store     cache.Store
}

// NewMachineLearningAlgorithmCache creates a new cache + resource watcher for MLA resources
func NewMachineLearningAlgorithmCache(client *nexuscore.Clientset, resourceNamespace string, logger klog.Logger) *MachineLearningAlgorithmCache {
	factory := nexusinf.NewSharedInformerFactoryWithOptions(client, time.Second*30, nexusinf.WithNamespace(resourceNamespace))
	watcher := factory.Science().V1().MachineLearningAlgorithms()

	return &MachineLearningAlgorithmCache{
		logger:    logger,
		factory:   factory,
		watcher:   watcher,
		lister:    watcher.Lister(),
		hasSynced: watcher.Informer().HasSynced,
		store:     watcher.Informer().GetStore(),
	}
}

// Init starts informers and sync the cache
func (c *MachineLearningAlgorithmCache) Init(ctx context.Context) error {
	// Set up an event handler for when Machine Learning Algorithm resources change
	_, err := c.watcher.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onConfigurationAdded,
		UpdateFunc: c.onConfigurationUpdated,
		DeleteFunc: c.onConfigurationDeleted,
	})

	if err != nil {
		return err
	}

	c.factory.Start(ctx.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), c.hasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	c.logger.Info("Resource informers synced")

	return nil
}

func (c *MachineLearningAlgorithmCache) onConfigurationAdded(obj interface{}) {
	objectRef, err := cache.ObjectToName(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.logger.V(3).Info("New configuration loaded", "algorithm", objectRef.Name)
}

func (c *MachineLearningAlgorithmCache) onConfigurationUpdated(old, new interface{}) {
	_, oldErr := cache.ObjectToName(old)
	newRef, newErr := cache.ObjectToName(new)

	if oldErr != nil {
		utilruntime.HandleError(oldErr)
		return
	}

	if newErr != nil {
		utilruntime.HandleError(newErr)
		return
	}

	c.logger.V(3).Info("Configuration updated", "algorithm", newRef.Name, "diff", diff.ObjectGoPrintSideBySide(old, new))
}
func (c *MachineLearningAlgorithmCache) onConfigurationDeleted(obj interface{}) {
	c.logger.V(3).Info("Configuration deleted", "algorithm", obj.(v1.MachineLearningAlgorithm).Name)
}

// GetConfiguration retrieves a cached MLA resource from informer cache
func (c *MachineLearningAlgorithmCache) GetConfiguration(algorithmName string) (*v1.MachineLearningAlgorithm, error) {
	config, exists, err := c.store.GetByKey(algorithmName)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	return config.(*v1.MachineLearningAlgorithm), nil
}
