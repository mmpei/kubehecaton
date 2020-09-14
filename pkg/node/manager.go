package node

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	clientV1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"registry/context"
	"k8s.io/klog"
	"fmt"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"github.com/mmpei/kubehecaton/pkg/common"
)

// NodeManager check the nodes related to, pick one by hash for pod scheduling
type NodeManager interface {
	GetBindNode(ctx context.Context, key string) (*v1.Node, error)
	GetHealthyNodes(ctx context.Context) ([]*v1.Node, error)
}

type nodeManager struct {
	lister clientV1.NodeLister
	synced cache.InformerSynced
}

func NewNodeManager(informerFactory informers.SharedInformerFactory) NodeManager {
	manager := &nodeManager{
		lister:     informerFactory.Core().V1().Nodes().Lister(),
		synced:     informerFactory.Core().V1().Nodes().Informer().HasSynced,
	}
	return manager
}

func (nm *nodeManager) GetBindNode(ctx context.Context, key string) (*v1.Node, error) {
	healthyNodes, err := nm.GetHealthyNodes(ctx)
	if err != nil {
		return nil, err
	}

	if len(healthyNodes) == 0 {
		return nil, fmt.Errorf("no node in healthy status: %s", key)
	}

	// calculate hash and pick one for pod scheduling
	node := nm.populate(healthyNodes, key)
	checkNodeStatus(node)
	return node, nil
}

func (nm *nodeManager) GetHealthyNodes(ctx context.Context) ([]*v1.Node, error) {
	nm.checkInit(ctx)

	var healthyNodes []*v1.Node
	requirement, err := labels.NewRequirement(common.LabelNodeCodeServer, selection.Exists, []string{})
	if err != nil {
		return nil, fmt.Errorf("invalid node label selector")
	}
	nodes, err := nm.lister.List(labels.NewSelector().Add(*requirement))
	if err != nil {
		return nil, fmt.Errorf("list node error: %v", err)
	}

	for _, node := range nodes {
		for _, condition := range node.Status.Conditions {
			if condition.Type == v1.NodeReady && condition.Status == v1.ConditionTrue {
				healthyNodes = append(healthyNodes, node)
			}
		}
	}
	return healthyNodes, nil
}

func (nm *nodeManager) checkInit(ctx context.Context) {
	if nm.synced() {
		return
	}
	if ok := cache.WaitForCacheSync(ctx.Done(), nm.synced); !ok {
		klog.Fatalf("failed to wait for node caches to sync")
		return
	}
}

func (nm *nodeManager) populate(nodes []*v1.Node, key string) *v1.Node {
	if len(nodes) == 0 {
		return nil
	}
	nmap := make(map[string]*v1.Node, 0)
	var nn []string
	for _, n := range nodes {
		nmap[n.Name] = n
		nn = append(nn, n.Name)
	}

	nn = common.SortByWeight(nn, []byte(key))
	return nmap[nn[0]]
}

func checkNodeStatus(node *v1.Node) {
	if node == nil {
		return
	}

	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case v1.NodeOutOfDisk, v1.NodeNetworkUnavailable:
			if condition.Status == v1.ConditionTrue {
				klog.Errorf("scheduling target node error: %s  %s ", node.Name, condition.Type)
			}
		case v1.NodeMemoryPressure, v1.NodeDiskPressure, v1.NodePIDPressure:
			if condition.Status == v1.ConditionTrue {
				klog.Warningf("scheduling target node warning: %s  %s ", node.Name, condition.Type)
			}
		}
	}
}
