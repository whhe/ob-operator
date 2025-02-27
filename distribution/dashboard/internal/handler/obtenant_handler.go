package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/oceanbase/oceanbase-dashboard/internal/business/oceanbase"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/param"
	"github.com/oceanbase/oceanbase-dashboard/internal/model/response"
	httpErr "github.com/oceanbase/oceanbase-dashboard/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	logger "github.com/sirupsen/logrus"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// @ID ListAllTenants
// @Tags Obtenant
// @Summary List all tenants
// @Description List all tenants and return them
// @Accept application/json
// @Produce application/json
// @Param obcluster query string false "obcluster to filter"
// @Success 200 object response.APIResponse{data=[]response.OBTenantBrief}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [GET]
// @Security ApiKeyAuth
func ListAllTenants(c *gin.Context) ([]*response.OBTenantBrief, error) {
	selector := ""
	if c.Query("obcluster") != "" {
		selector = fmt.Sprintf("ref-obcluster=%s", c.Query("obcluster"))
	}
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	tenants, err := oceanbase.ListAllOBTenants(c, listOptions)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenants, nil
}

// @ID GetTenant
// @Tags Obtenant
// @Summary Get tenant
// @Description Get an obtenant in a specific namespace
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name} [GET]
// @Security ApiKeyAuth
func GetTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.GetOBTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound(err.Error())
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID CreateTenant
// @Tags Obtenant
// @Summary Create tenant
// @Description Create an obtenant in a specific namespace, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateOBTenantParam true "create obtenant request body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants [PUT]
// @Security ApiKeyAuth
func CreateTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	tenantParam := &param.CreateOBTenantParam{}
	err := c.BindJSON(tenantParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logger.Infof("Create obtenant: %+v", tenantParam)
	tenant, err := oceanbase.CreateOBTenant(c, types.NamespacedName{
		Namespace: tenantParam.Name,
		Name:      tenantParam.Namespace,
	}, tenantParam)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID DeleteTenant
// @Tags Obtenant
// @Summary Delete tenant
// @Description Delete an obtenant in a specific namespace, ask user to confrim the deletion carefully
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name} [DELETE]
// @Security ApiKeyAuth
func DeleteTenant(c *gin.Context) (interface{}, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteOBTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		if kubeerrors.IsNotFound(err) {
			return nil, httpErr.NewNotFound(err.Error())
		}
		return nil, httpErr.NewInternal(err.Error())
	}
	return nil, nil
}

// @Deprecated: use PatchTenant instead
func ModifyUnitNumber(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	unitNumberParam := &param.ModifyUnitNumber{}
	err = c.BindJSON(unitNumberParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.ModifyOBTenantUnitNumber(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, unitNumberParam.UnitNumber)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @Deprecated: use PatchTenant instead
func ModifyUnitConfig(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := struct {
		Name      string `uri:"name"`
		Namespace string `uri:"namespace"`
		Zone      string `uri:"zone"`
	}{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	unitConfig := param.UnitConfig{}
	err = c.BindJSON(&unitConfig)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.ModifyOBTenantUnitConfig(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, nn.Zone, &unitConfig)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID PatchTenant
// @Tags Obtenant
// @Summary Patch tenant's configuration
// @Description Patch tenant's configuration
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.PatchTenant true "patch tenant body"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name} [PATCH]
// @Security ApiKeyAuth
func PatchTenant(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := param.NamespacedName{}
	err := c.BindUri(&nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	patch := param.PatchTenant{}
	err = c.BindJSON(&patch)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if patch.UnitNumber == nil && patch.UnitConfig == nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.PatchTenant(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, &patch)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// @ID ChangeUserPassword
// @Tags Obtenant
// @Summary Change root password of specific tenant
// @Description Change root password of specific tenant, encrypted by AES
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ChangeUserPassword true "new password"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name}/userCredentials [POST]
// @Security ApiKeyAuth
func ChangeUserPassword(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	passwordParam := &param.ChangeUserPassword{}
	err = c.BindJSON(passwordParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if passwordParam.User != "root" {
		return nil, httpErr.NewBadRequest("only root user is supported")
	}
	tenant, err := oceanbase.ModifyOBTenantRootPassword(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	}, passwordParam.Password)

	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID ReplayStandbyLog
// @Tags Obtenant
// @Summary Replay standby log of specific standby tenant
// @Description Replay standby log of specific standby tenant
// @Accept application/json
// @Produce application/json
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ReplayStandbyLog true "target timestamp to replay to"
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Router /api/v1/obtenants/{namespace}/{name}/logreplay [POST]
// @Security ApiKeyAuth
func ReplayStandbyLog(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	logReplayParam := &param.ReplayStandbyLog{}
	err = c.BindJSON(logReplayParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	if logReplayParam.Timestamp == nil {
		return nil, httpErr.NewBadRequest("timestamp is required")
	}
	tenant, err := oceanbase.ReplayStandbyLog(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, *logReplayParam.Timestamp)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID UpgradeTenantVersion
// @Tags Obtenant
// @Summary Upgrade tenant compatibility version of specific tenant
// @Description Upgrade tenant compatibility version of specific tenant to match the version of cluster
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/version [POST]
// @Security ApiKeyAuth
func UpgradeTenantVersion(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.UpgradeTenantVersion(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return tenant, nil
}

// @ID ChangeTenantRole
// @Tags Obtenant
// @Summary Change tenant role of specific tenant
// @Description Change tenant role of specific tenant, if a tenant is a standby tenant, it can be changed to primary tenant, vice versa
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.OBTenantDetail}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.ChangeTenantRole true "target role to change to"
// @Router /api/v1/obtenants/{namespace}/{name}/role [POST]
// @Security ApiKeyAuth
func ChangeTenantRole(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	p := param.ChangeTenantRole{}
	err = c.BindJSON(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	tenant, err := oceanbase.ChangeTenantRole(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, &p)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// @ID CreateBackupPolicy
// @Tags Obtenant
// @Summary Create backup policy of specific tenant
// @Description Create backup policy of specific tenant, passwords should be encrypted by AES
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.CreateBackupPolicy true "create backup policy request body"
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [PUT]
// @Security ApiKeyAuth
func CreateBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	createPolicyParam := &param.CreateBackupPolicy{}
	err = c.BindJSON(createPolicyParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	policy, err := oceanbase.CreateTenantBackupPolicy(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, createPolicyParam)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// @ID UpdateBackupPolicy
// @Tags Obtenant
// @Summary Update backup policy of specific tenant
// @Description Update backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param body body param.UpdateBackupPolicy true "update backup policy request body"
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [POST]
// @Security ApiKeyAuth
func UpdateBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	updatePolicyParam := &param.UpdateBackupPolicy{}
	err = c.BindJSON(updatePolicyParam)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	policy, err := oceanbase.UpdateTenantBackupPolicy(c, types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}, updatePolicyParam)
	if err != nil {
		return nil, err
	}
	return policy, nil
}

// @ID DeleteBackupPolicy
// @Tags Obtenant
// @Summary Delete backup policy of specific tenant
// @Description Delete backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [DELETE]
// @Security ApiKeyAuth
func DeleteBackupPolicy(c *gin.Context) (*response.OBTenantDetail, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	err = oceanbase.DeleteTenantBackupPolicy(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return nil, nil
}

// @ID GetBackupPolicy
// @Tags Obtenant
// @Summary Get backup policy of specific tenant
// @Description Get backup policy of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=response.BackupPolicy}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Router /api/v1/obtenants/{namespace}/{name}/backupPolicy [GET]
// @Security ApiKeyAuth
func GetBackupPolicy(c *gin.Context) (*response.BackupPolicy, error) {
	nn := &param.NamespacedName{}
	err := c.BindUri(nn)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	policy, err := oceanbase.GetTenantBackupPolicy(c, types.NamespacedName{
		Namespace: nn.Namespace,
		Name:      nn.Name,
	})
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return policy, nil
}

// @ID ListBackupJobs
// @Tags Obtenant
// @Summary List backup jobs of specific tenant
// @Description List backup jobs of specific tenant
// @Accept application/json
// @Produce application/json
// @Success 200 object response.APIResponse{data=[]response.BackupJob}
// @Failure 400 object response.APIResponse
// @Failure 401 object response.APIResponse
// @Failure 500 object response.APIResponse
// @Param namespace path string true "obtenant namespace"
// @Param name path string true "obtenant name"
// @Param type path string true "backup job type" Enums(FULL,INCR,CLEAN,ARCHIVE)
// @Param limit query int false "limit" default(10)
// @Router /api/v1/obtenants/{namespace}/{name}/backup/{type}/jobs [GET]
// @Security ApiKeyAuth
func ListBackupJobs(c *gin.Context) ([]*response.BackupJob, error) {
	p := struct {
		Namespace string `uri:"namespace"`
		Name      string `uri:"name"`
		Type      string `uri:"type"`
	}{}
	err := c.BindUri(&p)
	if err != nil {
		return nil, httpErr.NewBadRequest(err.Error())
	}
	limit := 10
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			return nil, httpErr.NewBadRequest(err.Error())
		}
	}
	jobs, err := oceanbase.ListBackupJobs(c, types.NamespacedName{
		Namespace: p.Namespace,
		Name:      p.Name,
	}, p.Type, limit)
	if err != nil {
		return nil, httpErr.NewInternal(err.Error())
	}
	return jobs, nil
}
