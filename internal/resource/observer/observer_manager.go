/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package observer

import (
	"context"
	"strings"

	"github.com/oceanbase/ob-operator/internal/telemetry"
	"github.com/oceanbase/ob-operator/pkg/oceanbase-sdk/model"
	taskstatus "github.com/oceanbase/ob-operator/pkg/task/const/status"
	"github.com/oceanbase/ob-operator/pkg/task/const/strategy"

	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
	"sigs.k8s.io/controller-runtime/pkg/client"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	v1alpha1 "github.com/oceanbase/ob-operator/api/v1alpha1"
	clusterstatus "github.com/oceanbase/ob-operator/internal/const/status/obcluster"
	serverstatus "github.com/oceanbase/ob-operator/internal/const/status/observer"
	resourceutils "github.com/oceanbase/ob-operator/internal/resource/utils"
	opresource "github.com/oceanbase/ob-operator/pkg/coordinator"
	"github.com/oceanbase/ob-operator/pkg/task"
	tasktypes "github.com/oceanbase/ob-operator/pkg/task/types"
)

type OBServerManager struct {
	opresource.ResourceManager
	Ctx      context.Context
	OBServer *v1alpha1.OBServer
	Client   client.Client
	Recorder telemetry.Recorder
	Logger   *logr.Logger
}

func (m *OBServerManager) GetTaskFunc(name tasktypes.TaskName) (tasktypes.TaskFunc, error) {
	switch name {
	case tCreateOBPVC:
		return m.CreateOBPVC, nil
	case tCreateOBPod:
		return m.CreateOBPod, nil
	case tWaitOBServerReady:
		return m.WaitOBServerReady, nil
	case tWaitOBClusterBootstrapped:
		return m.WaitOBClusterBootstrapped, nil
	case tAddServer:
		return m.AddServer, nil
	case tDeleteOBServerInCluster:
		return m.DeleteOBServerInCluster, nil
	case tWaitOBServerDeletedInCluster:
		return m.WaitOBServerDeletedInCluster, nil
	case tWaitOBServerPodReady:
		return m.WaitOBServerPodReady, nil
	case tWaitOBServerActiveInCluster:
		return m.WaitOBServerActiveInCluster, nil
	case tUpgradeOBServerImage:
		return m.UpgradeOBServerImage, nil
	case tAnnotateOBServerPod:
		return m.AnnotateOBServerPod, nil
	case tDeletePod:
		return m.DeletePod, nil
	case tWaitForPodDeleted:
		return m.WaitForPodDeleted, nil
	case tExpandPVC:
		return m.ResizePVC, nil
	case tWaitForPVCResized:
		return m.WaitForPVCResized, nil
	default:
		return nil, errors.Errorf("Can not find an function for task %s", name)
	}
}

func (m *OBServerManager) IsNewResource() bool {
	return m.OBServer.Status.Status == ""
}

func (m *OBServerManager) GetStatus() string {
	return m.OBServer.Status.Status
}

func (m *OBServerManager) InitStatus() {
	m.Logger.Info("newly created server, init status")
	status := v1alpha1.OBServerStatus{
		Image:  m.OBServer.Spec.OBServerTemplate.Image,
		Status: serverstatus.New,
	}
	m.OBServer.Status = status
}

func (m *OBServerManager) SetOperationContext(c *tasktypes.OperationContext) {
	m.OBServer.Status.OperationContext = c
}

func (m *OBServerManager) SupportStaticIp() bool {
	switch m.OBServer.Status.CNI {
	case oceanbaseconst.CNICalico:
		return true
	default:
		return false
	}
}

func (m *OBServerManager) getCurrentOBServerFromOB() (*model.OBServer, error) {
	if m.OBServer.Status.PodIp == "" {
		err := errors.New("pod ip is empty")
		m.Logger.Error(err, "unable to get observer info")
		return nil, err
	}
	observerInfo := &model.ServerInfo{
		Ip:   m.OBServer.Status.PodIp,
		Port: oceanbaseconst.RpcPort,
	}
	mode, modeExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if modeExist && mode == oceanbaseconst.ModeStandalone {
		observerInfo.Ip = "127.0.0.1"
	}
	operationManager, err := m.getOceanbaseOperationManager()
	if err != nil {
		return nil, errors.Wrapf(err, "Get oceanbase operation manager failed")
	}
	return operationManager.GetServer(observerInfo)
}

func (m *OBServerManager) retryUpdateStatus() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		observer, err := m.getOBServer()
		if err != nil {
			return client.IgnoreNotFound(err)
		}
		observer.Status = *m.OBServer.Status.DeepCopy()
		return m.Client.Status().Update(m.Ctx, observer)
	})
}

func (m *OBServerManager) setRecoveryStatus() {
	mode, modeExist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode)
	if m.SupportStaticIp() || (modeExist && mode == oceanbaseconst.ModeStandalone) {
		m.Logger.Info("current cni supports specific static ip address or the cluster runs as standalone, recover by recreate pod")
		m.OBServer.Status.Status = serverstatus.Recover
	} else {
		m.Logger.Info("observer not recoverable, delete current observer and wait recreate")
		m.OBServer.Status.Status = serverstatus.Unrecoverable
	}
}

func (m *OBServerManager) UpdateStatus() error {
	// update deleting status when object is deleting
	if m.IsDeleting() {
		m.OBServer.Status.Status = serverstatus.Deleting
	} else if m.OBServer.Status.Status == "Failed" {
		return nil
	} else {
		// get Pod status and update
		pod, err := m.getPod()
		if err != nil {
			if kubeerrors.IsNotFound(err) {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("pod not found")
				if m.OBServer.Status.Status == serverstatus.Running {
					m.setRecoveryStatus()
				}
			} else {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("observer status is not running, wait task finish")
				return errors.Wrap(err, "get pod when update status")
			}
		} else {
			m.OBServer.Status.Ready = apipod.IsPodReady(pod)
			m.OBServer.Status.PodPhase = pod.Status.Phase
			m.OBServer.Status.PodIp = pod.Status.PodIP
			m.OBServer.Status.NodeIp = pod.Status.HostIP
			// TODO update from obcluster
			m.OBServer.Status.CNI = resourceutils.GetCNIFromAnnotation(pod)
		}
		pvcs, err := m.getPVCs()
		if err != nil {
			m.Logger.Info("get pvc failed: " + err.Error())
		}
		// 1. Check status of observer in OB database
		if m.OBServer.Status.Status == serverstatus.Running {
			m.Logger.V(oceanbaseconst.LogLevelDebug).Info("check observer in obcluster")
			observer, err := m.getCurrentOBServerFromOB()
			if err != nil {
				m.Logger.V(oceanbaseconst.LogLevelDebug).Info("Get observer failed, check next time")
			} else if observer == nil {
				m.OBServer.Status.Status = serverstatus.AddServer
			} else if mode, exist := resourceutils.GetAnnotationField(m.OBServer, oceanbaseconst.AnnotationsMode); exist && mode == oceanbaseconst.ModeStandalone {
				if pod.Spec.Containers[0].Resources.Limits.Cpu().Cmp(m.OBServer.Spec.OBServerTemplate.Resource.Cpu) != 0 ||
					pod.Spec.Containers[0].Resources.Limits.Memory().Cmp(m.OBServer.Spec.OBServerTemplate.Resource.Memory) != 0 {
					m.OBServer.Status.Status = serverstatus.ScaleUp
				}
			} else if pvcs != nil && len(pvcs.Items) > 0 && m.checkIfStorageExpand(pvcs) {
				m.OBServer.Status.Status = serverstatus.ExpandPVC
			}
		}

		// 2. Check CNI Annotations and upgrade
		if m.OBServer.Status.Status == serverstatus.Running {
			if resourceutils.NeedAnnotation(pod, m.OBServer.Status.CNI) {
				m.OBServer.Status.Status = serverstatus.Annotate
			} else {
				for _, container := range pod.Spec.Containers {
					if container.Name == oceanbaseconst.ContainerName {
						m.OBServer.Status.Image = container.Image
						break
					}
				}
				if m.OBServer.Spec.OBServerTemplate.Image != m.OBServer.Status.Image {
					m.Logger.Info("Found image changed, begin upgrade")
					m.OBServer.Status.Status = serverstatus.Upgrade
				}
			}
		}

		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("update observer status", "status", m.OBServer.Status)
	}

	err := m.retryUpdateStatus()
	if err != nil {
		m.Logger.Error(err, "Got error when update observer status")
	}
	return err
}

func (m *OBServerManager) IsDeleting() bool {
	return !m.OBServer.ObjectMeta.DeletionTimestamp.IsZero()
}

func (m *OBServerManager) CheckAndUpdateFinalizers() error {
	finalizerFinished := false
	obcluster, err := m.getOBCluster()
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			m.Logger.Info("OBCluster is deleted, no need to wait finalizer")
			finalizerFinished = true
		} else {
			m.Logger.Error(err, "query obcluster failed")
			return errors.Wrap(err, "Get obcluster failed")
		}
	} else if !obcluster.ObjectMeta.DeletionTimestamp.IsZero() {
		m.Logger.Info("OBCluster is deleting, no need to wait finalizer")
		finalizerFinished = true
	} else {
		finalizerFinished = m.OBServer.Status.Status == serverstatus.FinalizerFinished
	}
	if finalizerFinished {
		m.Logger.Info("Finalizer finished")
		m.OBServer.ObjectMeta.Finalizers = make([]string, 0)
		err := m.Client.Update(m.Ctx, m.OBServer)
		if err != nil {
			m.Logger.Error(err, "update observer instance failed")
			return errors.Wrapf(err, "Update observer %s in K8s failed", m.OBServer.Name)
		}
	}
	return nil
}

func (m *OBServerManager) GetTaskFlow() (*tasktypes.TaskFlow, error) {
	// exists unfinished task flow, return the last task flow
	if m.OBServer.Status.OperationContext != nil {
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("get task flow from observer status")
		return tasktypes.NewTaskFlow(m.OBServer.Status.OperationContext), nil
	}
	// newly created observer
	var taskFlow *tasktypes.TaskFlow
	var err error
	var obcluster *v1alpha1.OBCluster

	m.Logger.V(oceanbaseconst.LogLevelTrace).Info("create task flow according to observer status")
	switch m.OBServer.Status.Status {
	case serverstatus.New:
		obcluster, err = m.getOBCluster()
		if err != nil {
			return nil, errors.Wrap(err, "Get obcluster")
		}
		if obcluster.Status.Status == clusterstatus.New {
			// created when create obcluster
			m.Logger.Info("Create observer when create obcluster")
			taskFlow, err = task.GetRegistry().Get(fPrepareOBServerForBootstrap)
		} else {
			// created normally
			m.Logger.Info("Create observer when obcluster already exists")
			taskFlow, err = task.GetRegistry().Get(fCreateOBServer)
		}
		if err != nil {
			return nil, errors.Wrap(err, "Get create observer task flow")
		}
	case serverstatus.BootstrapReady:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when bootstrap ready")
		taskFlow, err = task.GetRegistry().Get(fMaintainOBServerAfterBootstrap)
	case serverstatus.Deleting:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer deleting")
		taskFlow, err = task.GetRegistry().Get(fDeleteOBServerFinalizer)
	case serverstatus.Upgrade:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer upgrade")
		taskFlow, err = task.GetRegistry().Get(fUpgradeOBServer)
	case serverstatus.Recover:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer need recover")
		taskFlow, err = task.GetRegistry().Get(fRecoverOBServer)
	case serverstatus.Annotate:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer need set annotation")
		taskFlow, err = task.GetRegistry().Get(fAnnotateOBServerPod)
	case serverstatus.AddServer:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer need to be added to obcluster")
		taskFlow, err = task.GetRegistry().Get(fAddServerInOB)
	case serverstatus.ScaleUp:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer need to be scaled up")
		taskFlow, err = task.GetRegistry().Get(fScaleUpOBServer)
	case serverstatus.ExpandPVC:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("Get task flow when observer need to expand pvc")
		taskFlow, err = task.GetRegistry().Get(fExpandPVC)
	default:
		m.Logger.V(oceanbaseconst.LogLevelTrace).Info("no need to run anything for observer")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if taskFlow.OperationContext.OnFailure.Strategy == "" {
		taskFlow.OperationContext.OnFailure.Strategy = strategy.StartOver
		if taskFlow.OperationContext.OnFailure.NextTryStatus == "" {
			taskFlow.OperationContext.OnFailure.NextTryStatus = serverstatus.Running
		}
	}
	return taskFlow, nil
}

func (m *OBServerManager) ClearTaskInfo() {
	m.OBServer.Status.Status = serverstatus.Running
	m.OBServer.Status.OperationContext = nil
}

func (m *OBServerManager) FinishTask() {
	m.OBServer.Status.Status = m.OBServer.Status.OperationContext.TargetStatus
	m.OBServer.Status.OperationContext = nil
}

func (m *OBServerManager) HandleFailure() {
	if m.IsDeleting() {
		m.OBServer.Status.Status = serverstatus.Deleting
		m.OBServer.Status.OperationContext = nil
	} else {
		operationContext := m.OBServer.Status.OperationContext
		failureRule := operationContext.OnFailure
		switch failureRule.Strategy {
		case strategy.StartOver:
			if m.OBServer.Status.Status != failureRule.NextTryStatus {
				m.OBServer.Status.Status = failureRule.NextTryStatus
				m.OBServer.Status.OperationContext = nil
			} else {
				m.OBServer.Status.OperationContext.Idx = 0
				m.OBServer.Status.OperationContext.TaskStatus = ""
				m.OBServer.Status.OperationContext.TaskId = ""
				m.OBServer.Status.OperationContext.Task = ""
			}
		case strategy.RetryFromCurrent:
			operationContext.TaskStatus = taskstatus.Pending
		case strategy.Pause:
		}
	}
}

func (m *OBServerManager) PrintErrEvent(err error) {
	m.Recorder.Event(m.OBServer, corev1.EventTypeWarning, "task exec failed", err.Error())
}

func (m *OBServerManager) generateNamespacedName(name string) types.NamespacedName {
	var namespacedName types.NamespacedName
	namespacedName.Namespace = m.OBServer.Namespace
	namespacedName.Name = name
	return namespacedName
}

func (m *OBServerManager) getPod() (*corev1.Pod, error) {
	// this label always exists
	pod := &corev1.Pod{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), pod)
	if err != nil {
		return nil, errors.Wrap(err, "get pod")
	}
	return pod, nil
}

func (m *OBServerManager) getOBCluster() (*v1alpha1.OBCluster, error) {
	// this label always exists
	clusterName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBCluster]
	obcluster := &v1alpha1.OBCluster{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(clusterName), obcluster)
	if err != nil {
		return nil, errors.Wrap(err, "get obcluster")
	}
	return obcluster, nil
}

// get observer from K8s api server
func (m *OBServerManager) getOBServer() (*v1alpha1.OBServer, error) {
	// this label always exists
	observer := &v1alpha1.OBServer{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(m.OBServer.Name), observer)
	if err != nil {
		return nil, errors.Wrap(err, "get observer")
	}
	return observer, nil
}

func (m *OBServerManager) getOBZone() (*v1alpha1.OBZone, error) {
	// this label always exists
	zoneName, _ := m.OBServer.Labels[oceanbaseconst.LabelRefOBZone]
	obzone := &v1alpha1.OBZone{}
	err := m.Client.Get(m.Ctx, m.generateNamespacedName(zoneName), obzone)
	if err != nil {
		return nil, errors.Wrap(err, "get obzone")
	}
	return obzone, nil
}

func (m *OBServerManager) ArchiveResource() {
	m.Logger.Info("Archive observer", "observer", m.OBServer.Name)
	m.Recorder.Event(m.OBServer, "Archive", "", "archive observer")
	m.OBServer.Status.Status = "Failed"
	m.OBServer.Status.OperationContext = nil
}

func (m *OBServerManager) getPVCs() (*corev1.PersistentVolumeClaimList, error) {
	pvcs := &corev1.PersistentVolumeClaimList{}
	err := m.Client.List(m.Ctx, pvcs, client.InNamespace(m.OBServer.Namespace), client.MatchingLabels{oceanbaseconst.LabelRefUID: m.OBServer.Labels[oceanbaseconst.LabelRefUID]})
	if err != nil {
		return nil, errors.Wrap(err, "list pvc")
	}
	return pvcs, nil
}

func (m *OBServerManager) checkIfStorageExpand(pvcs *corev1.PersistentVolumeClaimList) bool {
	for _, pvc := range pvcs.Items {
		switch {
		case strings.HasSuffix(pvc.Name, oceanbaseconst.DataVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.ClogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size) < 0 {
				return true
			}
		case strings.HasSuffix(pvc.Name, oceanbaseconst.LogVolumeSuffix):
			if pvc.Spec.Resources.Requests.Storage().Cmp(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size) < 0 {
				return true
			}
		case pvc.Name == m.OBServer.Name:
			sum := resource.Quantity{}
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.DataStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.RedoLogStorage.Size)
			sum.Add(m.OBServer.Spec.OBServerTemplate.Storage.LogStorage.Size)
			if pvc.Spec.Resources.Requests.Storage().Cmp(sum) < 0 {
				return true
			}
		}
	}
	return false
}
