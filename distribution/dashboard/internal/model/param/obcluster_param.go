package param

import "github.com/oceanbase/oceanbase-dashboard/internal/model/common"

type ZoneTopology struct {
	Zone         string          `json:"zone"`
	Replicas     int             `json:"replicas"`
	NodeSelector []common.KVPair `json:"nodeSelector,omitempty"`
}

type OBServerStorageSpec struct {
	Data    common.StorageSpec `json:"data"`
	RedoLog common.StorageSpec `json:"redoLog"`
	Log     common.StorageSpec `json:"log"`
}

type MonitorStorageSpec struct {
	Config common.StorageSpec `json:"config"`
}

type OBServerSpec struct {
	Image    string               `json:"image"`
	Resource common.ResourceSpec  `json:"resource"`
	Storage  *OBServerStorageSpec `json:"storage"`
}

type MonitorSpec struct {
	Image    string              `json:"image"`
	Resource common.ResourceSpec `json:"resource"`
}

type NFSVolumeSpec struct {
	Address string `json:"address"`
	Path    string `json:"path"`
}

type CreateOBClusterParam struct {
	Namespace    string          `json:"namespace"`
	Name         string          `json:"name"`
	ClusterName  string          `json:"clusterName"`
	ClusterId    int64           `json:"clusterId"`
	RootPassword string          `json:"rootPassword"`
	Topology     []ZoneTopology  `json:"topology"`
	OBServer     *OBServerSpec   `json:"observer"`
	Monitor      *MonitorSpec    `json:"monitor"`
	Parameters   []common.KVPair `json:"parameters"`
	BackupVolume *NFSVolumeSpec  `json:"backupVolume"`
}

type UpgradeOBClusterParam struct {
	Image string `json:"image"`
}

type ScaleOBServerParam struct {
	Replicas int `json:"replicas"`
}

type K8sObjectIdentity struct {
	Namespace string `json:"namespace" uri:"namespace" binding:"required"`
	Name      string `json:"name" uri:"name" binding:"required"`
}

type OBZoneIdentity struct {
	Namespace  string `json:"namespace" uri:"namespace" binding:"required"`
	Name       string `json:"name" uri:"name" binding:"required"`
	OBZoneName string `json:"obzoneName" uri:"obzoneName" binding:"required"`
}
