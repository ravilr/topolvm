package scheduler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/topolvm/topolvm"
	corev1 "k8s.io/api/core/v1"
)

const gracePeriod = 200 * time.Second

func filterNodes(nodes corev1.NodeList, requested map[string]int64) ExtenderFilterResult {
	if len(requested) == 0 {
		return ExtenderFilterResult{
			Nodes: &nodes,
		}
	}

	failedNodes := make([]string, len(nodes.Items))
	wg := &sync.WaitGroup{}
	wg.Add(len(nodes.Items))
	for i := range nodes.Items {
		reason := &failedNodes[i]
		node := nodes.Items[i]
		go func() {
			*reason = filterNode(node, requested)
			wg.Done()
		}()
	}
	wg.Wait()
	result := ExtenderFilterResult{
		Nodes:       &corev1.NodeList{},
		FailedNodes: FailedNodesMap{},
	}
	for i, reason := range failedNodes {
		if len(reason) == 0 {
			result.Nodes.Items = append(result.Nodes.Items, nodes.Items[i])
		} else {
			result.FailedNodes[nodes.Items[i].Name] = reason
		}
	}
	return result
}

func filterNode(node corev1.Node, requested map[string]int64) string {
	for dc, required := range requested {
		val, ok := node.Annotations[topolvm.CapacityKeyPrefix+dc]
		if !ok {
			return "no capacity annotation"
		}
		capacity, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return "bad capacity annotation: " + val
		}
		if capacity < uint64(required) {
			return "out of VG free space"
		}
		t, ok := node.Annotations[topolvm.NodeHeartbeatKey]
		if !ok {
			return "no node probe timestamp"
		}
		ts, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return "bad node probe timestamp annotation: " + t
		}
		if time.Now().After(time.Unix(ts, 0).Add(gracePeriod)) {
			return "topolvm node probe timestamp was last set longer ago than 3 minutes: t"
		}
	}
	return ""
}

func extractRequestedSize(pod *corev1.Pod) map[string]int64 {
	result := make(map[string]int64)
	for k, v := range pod.Annotations {
		if !strings.HasPrefix(k, topolvm.CapacityKeyPrefix) {
			continue
		}
		dc := k[len(topolvm.CapacityKeyPrefix):]
		capacity, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		result[dc] = capacity
	}

	return result
}

func (s scheduler) predicate(w http.ResponseWriter, r *http.Request) {
	var input ExtenderArgs

	reader := http.MaxBytesReader(w, r.Body, 10<<20)
	err := json.NewDecoder(reader).Decode(&input)
	if err != nil || input.Nodes == nil || input.Pod == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	requested := extractRequestedSize(input.Pod)
	result := filterNodes(*input.Nodes, requested)
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(result)
}
