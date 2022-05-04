package mybalancedallocation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/YunruiSun/syr-scheduler/pkg/mybalancedallocation/filter"
)

const (
	Name = "mybalancedallocation"
)

var (
	_ framework.ScorePlugin     = &MyBalancedAllocation{}
	_ framework.ScoreExtensions = &MyBalancedAllocation{}

	scheme = runtime.NewScheme()
)

type MyBalancedAllocation struct {
	handle framework.Handle
	cache  cache.Cache
}

func (m *MyBalancedAllocation) Name() string {
	return Name
}

func New(_ runtime.Object, h framework.Handle) (framework.Plugin, error) {
	mgrConfig := ctrl.GetConfigOrDie()
	mgrConfig.QPS = 1000
	mgrConfig.Burst = 1000

	mgr, err := ctrl.NewManager(mgrConfig, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "",
		LeaderElection:     false,
		Port:               9443,
	})

	if err != nil {
		klog.Error(err)
		return nil, err
	}

	go func() {
		if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			klog.Error(err)
			panic(err)
		}
	}()

	scvCache := mgr.GetCache()

	if scvCache.WaitForCacheSync(context.TODO()) {
		return &MyBalancedAllocation{
			handle: h,
			cache:  scvCache,
		}, nil
	} else {
		return nil, errors.New("cache Not Sync! 1")
	}
}

func (m *MyBalancedAllocation) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	// Get Node info from prometheus
	node_cpu_frac, err := m.getNodeCpuFraction(nodeName)
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
	}
	node_mem_frac, err := m.getNodeMemFraction(nodeName)
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
	}
	node_net_frac, err := m.getNodeNetFraction(nodeName)
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
	}
	node_cpu_core_num, err := m.getNodeCpuCoreNum(nodeName)
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
	}
	node_mem_allo, err := m.getNodeMemAllocatable(nodeName)
	if err != nil {
		klog.Errorf("CalculateScore Error: %v", err)
	}
	node_p := 0.5*float64(node_cpu_core_num) + 0.3*float64(node_mem_allo)/1000000000 + 0.2
	node_l := 0.5*node_cpu_frac + 0.3*node_mem_frac + 0.2*node_net_frac
	nodeScore := int64((1 - node_l/node_p) * 100)
	return nodeScore, framework.NewStatus(framework.Success, "")
}

func (m *MyBalancedAllocation) NormalizeScore(_ context.Context, _ *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var (
		highest int64 = 0
		lowest        = scores[0].Score
	)

	for _, nodeSocre := range scores {
		if nodeSocre.Score < lowest {
			lowest = nodeSocre.Score
		}
		if nodeSocre.Score > highest {
			highest = nodeSocre.Score
		}
	}

	if highest == lowest {
		lowest--
	}

	// Set Range to [0-100]
	for i, nodeScore := range scores {
		scores[i].Score = (nodeScore.Score - lowest) * 100 / (highest - lowest)
		klog.Infof("Node: %v, Score: %v in Plugin: Yoda When scheduling Pod: %v/%v", scores[i].Name, scores[i].Score, pod.GetNamespace(), pod.GetName())
	}
	return nil
}

func (m *MyBalancedAllocation) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (m *MyBalancedAllocation) getNodeCpuFraction(nodeName string) (float64, error) {
	queryString := fmt.Sprintf("instance:node_cpu_utilisation:rate1m{instance=\"%s\"}", nodeName)
	r, err := http.Get(fmt.Sprintf("http://192.168.146.100:30090/api/v1/query?query=%s", queryString))
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	jsonString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	cpuFraction, err := filter.ParseDataToFloat(string(jsonString))
	if err != nil {
		return 0, err
	}
	return cpuFraction, nil
}

// prometheus获取节点内存利用率
func (m *MyBalancedAllocation) getNodeMemFraction(nodeName string) (float64, error) {
	queryString := fmt.Sprintf("instance:node_memory_utilisation:ratio{instance=\"%s\"}", nodeName)
	r, err := http.Get(fmt.Sprintf("http://192.168.146.100:30090/api/v1/query?query=%s", queryString))
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	jsonString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	memFraction, err := filter.ParseDataToFloat(string(jsonString))
	if err != nil {
		return 0, err
	}
	return memFraction, nil
}

// prometheus获取节点网络带宽指标
func (m *MyBalancedAllocation) getNodeNetFraction(nodeName string) (float64, error) {
	queryString := fmt.Sprintf("instance:node_network_receive_bytes_excluding_lo:rate1m{instance=\"%s\"}", nodeName)
	r, err := http.Get(fmt.Sprintf("http://192.168.146.100:30090/api/v1/query?query=%s", queryString))
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	jsonString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	netDownloadSpeed, err := filter.ParseDataToFloat(string(jsonString))
	if err != nil {
		return 0, err
	}
	return netDownloadSpeed / 10000000, nil
}

// prometheus获取节点内存总容量（认为是可分配内存）
func (m *MyBalancedAllocation) getNodeMemAllocatable(nodeName string) (int64, error) {
	queryString := fmt.Sprintf("kube_node_status_allocatable_memory_bytes{node=\"%s\"}", nodeName)
	r, err := http.Get(fmt.Sprintf("http://192.168.146.100:30090/api/v1/query?query=%s", queryString))
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	jsonString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	memAllocatable, err := filter.ParseDataToInt(string(jsonString))
	if err != nil {
		return 0, err
	}
	return memAllocatable, nil
}

// prometheus获取节点网络带宽（认为是主机最大下载速度）
func (m *MyBalancedAllocation) getNodeCpuCoreNum(nodeName string) (int64, error) {
	queryString := fmt.Sprintf("kube_node_status_allocatable_cpu_cores{node=\"%s\"}", nodeName)
	r, err := http.Get(fmt.Sprintf("http://192.168.146.100:30090/api/v1/query?query=%s", queryString))
	if err != nil {
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)
	jsonString, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, err
	}
	cpuCoreNum, err := filter.ParseDataToInt(string(jsonString))
	if err != nil {
		return 0, err
	}
	return cpuCoreNum, nil
}
