package node

import (
	"k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"fmt"
	"time"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/runtime"
	clientV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type NodeChecker struct {
	clientset kubernetes.Interface
	workqueue workqueue.RateLimitingInterface

	lister clientV1.NodeLister
	synced cache.InformerSynced

	nodes map[string]*v1.Node
	healthNodes map[string]*v1.Node
}

func NewNodeChecker(clientset kubernetes.Interface,
	informerFactory informers.SharedInformerFactory) *NodeChecker {
	checker := &NodeChecker{
		clientset: clientset,
		lister:     informerFactory.Core().V1().Nodes().Lister(),
		synced:     informerFactory.Core().V1().Nodes().Informer().HasSynced,
		workqueue: workqueue.NewNamedRateLimitingQueue(
			workqueue.DefaultControllerRateLimiter(),
			"NodeChecker"),
	}

	klog.Info("Setting up event handlers")
	// Set up an event handler for when node change
	informerFactory.Core().V1().Nodes().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: checker.enqueueNode,
		UpdateFunc: func(old, new interface{}) {
			checker.enqueueNode(new)
		},
	})
	return checker
}

// enqueueNode takes a node resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than node.
func (nc *NodeChecker) enqueueNode(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}

	nc.workqueue.Add(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (nc *NodeChecker) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer nc.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting NodeChecker")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, nc.synced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	//todo: Init Nodes related
	requirement, err := labels.NewRequirement("key", selection.Equals, []string{"value"})
	if err != nil {
		return fmt.Errorf("invalid node label selector")
	}
	nodes, err := nc.lister.List(labels.NewSelector().Add(*requirement))
	if err != nil {
		return fmt.Errorf("list node error: %v", err)
	}

	nc.healthNodes = make(map[string]*v1.Node, 0)
	for _, node := range nodes {
		nodeCopy := node.DeepCopy()
		nc.healthNodes[nodeCopy.Name] = nodeCopy
	}

	klog.Info("Starting workers")
	// Launch workers to process Node resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(nc.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (nc *NodeChecker) runWorker() {
	for nc.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (nc *NodeChecker) processNextWorkItem() bool {
	obj, shutdown := nc.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer nc.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			nc.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// node resource to be synced.
		if err := nc.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			nc.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		nc.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (nc *NodeChecker) syncHandler(key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	//namespace, name, err := cache.SplitMetaNamespaceKey(key)
	return nil
}