package response

import "github.com/oceanbase/oceanbase-dashboard/internal/model/common"

type K8sEvent struct {
	Namespace  string  `json:"namespace"`
	Type       string  `json:"type"`
	Count      int32   `json:"count"`
	FirstOccur float64 `json:"firstOccur"`
	LastSeen   float64 `json:"lastSeen"`
	Reason     string  `json:"reason"`
	Object     string  `json:"object"`
	Message    string  `json:"message"`
}

type K8sNodeCondition struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type K8sNodeInfo struct {
	Name       string             `json:"name"`
	Status     string             `json:"status"`
	Conditions []K8sNodeCondition `json:"conditions"`
	Roles      []string           `json:"roles"`
	Labels     []common.KVPair    `json:"labels"`
	Uptime     float64            `json:"uptime"`
	Version    string             `json:"version"`
	InternalIP string             `json:"internalIP"`
	ExternalIP string             `json:"externalIP"`
	OS         string             `json:"os"`
	Kernel     string             `json:"kernel"`
	CRI        string             `json:"cri"`
}

type K8sNodeResource struct {
	CpuTotal    float64 `json:"cpuTotal"`
	CpuUsed     float64 `json:"cpuUsed"`
	CpuFree     float64 `json:"cpuFree"`
	MemoryTotal float64 `json:"memoryTotal"`
	MemoryUsed  float64 `json:"memoryUsed"`
	MemoryFree  float64 `json:"memoryFree"`
}

type K8sNode struct {
	Info     *K8sNodeInfo     `json:"info"`
	Resource *K8sNodeResource `json:"resource"`
}

type Namespace struct {
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
}

type StorageClass struct {
	Name                 string          `json:"name"`
	Provisioner          string          `json:"provisioner"`
	ReclaimPolicy        string          `json:"reclaimPolicy"`
	VolumeBindingMode    string          `json:"volumeBindingMode"`
	AllowVolumeExpansion bool            `json:"allowVolumeExpansion"`
	MountOptions         []string        `json:"mountOptions,omitempty"`
	Parameters           []common.KVPair `json:"parameters,omitempty"`
}
