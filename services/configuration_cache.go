package services

import (
	"context"
	"fmt"
	v1 "github.com/SneaksAndData/nexus-core/pkg/apis/science/v1"
	nexuscore "github.com/SneaksAndData/nexus-core/pkg/generated/clientset/versioned"
	nexusinf "github.com/SneaksAndData/nexus-core/pkg/generated/informers/externalversions"
	"github.com/SneaksAndData/nexus-core/pkg/resolvers"
	"github.com/SneaksAndData/nexus-core/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// NewNexusResourceCache creates a new cache for Nexus resources
func NewNexusResourceCache(client nexuscore.Interface, resourceNamespace string, logger klog.Logger, resyncPeriod *time.Duration) *NexusResourceCache {
	defaultResyncPeriod := time.Second * 30
	factory := nexusinf.NewSharedInformerFactoryWithOptions(client, *util.CoalescePointer(resyncPeriod, &defaultResyncPeriod), nexusinf.WithNamespace(resourceNamespace))
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

	if err != nil { // coverage-ignore
		return err
	}

	_, werr := c.workgroupInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onConfigurationAdded,
		UpdateFunc: c.onConfigurationUpdated,
		DeleteFunc: c.onConfigurationDeleted,
	})

	if werr != nil { // coverage-ignore
		return werr
	}

	c.factory.Start(ctx.Done())

	if ok := cache.WaitForCacheSync(ctx.Done(), c.templateInformer.HasSynced, c.workgroupInformer.HasSynced); !ok { // coverage-ignore
		return fmt.Errorf("failed to wait for informer caches to sync")
	}

	c.logger.Info("custom resource informers synced")

	return nil
}

func (c *NexusResourceCache) onConfigurationAdded(obj interface{}) { // coverage-ignore
	objectRef, err := cache.ObjectToName(obj)
	if err != nil { // coverage-ignore
		utilruntime.HandleError(err)
		return
	}

	c.logger.V(3).Info("resource loaded", "resource", objectRef.Name)
}

func (c *NexusResourceCache) onConfigurationUpdated(old, new interface{}) {
	_, oldErr := cache.ObjectToName(old)
	newRef, newErr := cache.ObjectToName(new)

	if oldErr != nil { // coverage-ignore
		utilruntime.HandleError(oldErr)
		return
	}

	if newErr != nil { // coverage-ignore
		utilruntime.HandleError(newErr)
		return
	}

	c.logger.V(3).Info("resource updated", "resource", newRef.Name, "diff", diff.ObjectGoPrintSideBySide(old, new))
}

func (c *NexusResourceCache) onConfigurationDeleted(obj interface{}) {
	// attempt to read the object metadata
	if object, ok := obj.(metav1.Object); !ok {
		// check if object was deleted while we were not watching by attempting to get its tombstone info
		tombstone, deleted := obj.(cache.DeletedFinalStateUnknown)
		if !deleted {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object, invalid type", "type", fmt.Sprintf("%T", obj))
			return
		}
		// recover object data from the tombstone
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			// If the object value is not too big and does not contain sensitive information then
			// it may be useful to include it.
			utilruntime.HandleErrorWithContext(context.Background(), nil, "Error decoding object tombstone, invalid type", "type", fmt.Sprintf("%T", tombstone.Obj))
			return
		}

		c.logger.V(3).Info("resource deleted", "resource", object.GetName())
	} else {
		c.logger.V(3).Info("resource deleted", "resource", object.GetName())
	}
}

// GetAlgorithmConfiguration retrieves a cached NexusAlgorithmTemplate resource from informer cache
func (c *NexusResourceCache) GetAlgorithmConfiguration(algorithmName string) (*v1.NexusAlgorithmTemplate, error) {
	return resolvers.GetCachedObject[v1.NexusAlgorithmTemplate](algorithmName, c.prefix, c.templateInformer)
}

// GetWorkgroupConfiguration retrieves a cached NexusAlgorithmTemplate resource from informer cache
func (c *NexusResourceCache) GetWorkgroupConfiguration(workgroupName string) (*v1.NexusAlgorithmWorkgroup, error) {
	return resolvers.GetCachedObject[v1.NexusAlgorithmWorkgroup](workgroupName, c.prefix, c.workgroupInformer)
}
