package services

import (
	"context"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusinf "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	resolvers "github.com/SneaksAndData/nexus-core/pkg/resolvers"
	"k8s.io/apimachinery/pkg/util/diff"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"time"
)

type NexusResourceCache struct {
	logger            klog.Logger
	factory           nexusinf.SharedInformerFactory
	templateInformer  cache.SharedIndexInformer
	workgroupInformer cache.SharedIndexInformer
	prefix            string
}

// NewNexusResourceCache creates a new cache + resource watcher for MLA resources
func NewNexusResourceCache(client *nexuscore.Clientset, resourceNamespace string, logger klog.Logger) *NexusResourceCache {
	factory := nexusinf.NewSharedInformerFactoryWithOptions(client, time.Second*30, nexusinf.WithNamespace(resourceNamespace))
	watcher := factory.Science().V1().NexusAlgorithmTemplates()
	workgroupWatcher := factory.Science().V1().NexusAlgorithmWorkgroups()

	return &NexusResourceCache{
		logger:            logger,
		factory:           factory,
		templateInformer:  watcher.Informer(),
		workgroupInformer: workgroupWatcher.Informer(),
		prefix:            resourceNamespace,
	}
}

// Init starts informers and sync the cache
func (c *NexusResourceCache) Init(ctx context.Context) error {
	// Set up an event handler for when Machine Learning Algorithm resources change
	_, err := c.templateInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onConfigurationAdded,
		UpdateFunc: c.onConfigurationUpdated,
		DeleteFunc: c.onConfigurationDeleted,
	})

	if err != nil {
		return err
	}

	_, werr := c.workgroupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onConfigurationAdded,
		UpdateFunc: c.onConfigurationUpdated,
		DeleteFunc: c.onConfigurationDeleted,
	})

	if werr != nil {
		return werr
	}

	c.factory.Start(ctx.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), c.templateInformer.HasSynced, c.workgroupInformer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for informer caches to sync")
	}

	c.logger.Info("custom resource informers synced")

	return nil
}

func (c *NexusResourceCache) onConfigurationAdded(obj interface{}) {
	objectRef, err := cache.ObjectToName(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.logger.V(3).Info("resource loaded", "algorithm", objectRef.Name)
}

func (c *NexusResourceCache) onConfigurationUpdated(old, new interface{}) {
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

	c.logger.V(3).Info("resource updated", "resource", newRef.Name, "diff", diff.ObjectGoPrintSideBySide(old, new))
}
func (c *NexusResourceCache) onConfigurationDeleted(obj interface{}) {
	c.logger.V(3).Info("resource deleted", "resource", obj.(*v1.NexusAlgorithmTemplate).Name)
}

// GetAlgorithmConfiguration retrieves a cached NexusAlgorithmTemplate resource from informer cache
func (c *NexusResourceCache) GetAlgorithmConfiguration(algorithmName string) (*v1.NexusAlgorithmTemplate, error) {
	return resolvers.GetCachedObject[v1.NexusAlgorithmTemplate](algorithmName, c.prefix, c.templateInformer)
}

// GetWorkgroupConfiguration retrieves a cached NexusAlgorithmTemplate resource from informer cache
func (c *NexusResourceCache) GetWorkgroupConfiguration(workgroupName string) (*v1.NexusAlgorithmWorkgroup, error) {
	return resolvers.GetCachedObject[v1.NexusAlgorithmWorkgroup](workgroupName, c.prefix, c.workgroupInformer)
}
