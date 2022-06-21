package upgrade

import (
	"encoding/json"
	"errors"
	"fmt"
	"k8s-installer/pkg/backup_restore"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/control_manager"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/docker"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/network"
	"k8s-installer/pkg/task_breaker"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"

	"github.com/emicklei/go-restful"
	"github.com/google/uuid"
)

func ApplyUpgradePlan(request *restful.Request, response *restful.Response) {
	planId := request.PathParameter("plan-id")
	if strings.TrimSpace(planId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "plan id"))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	plan, err := runtimeCache.GetUpgradePlan(planId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get upgrade plan detail due to error: %s", err.Error()))
		return
	} else if plan == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find upgrade plan %s", planId))
		return
	} else if plan.Status == constants.StatusError {
		utils.ResponseError(response, http.StatusPreconditionFailed, "This upgrade plan currently in error status. Please find the last operation and retry")
		return
	}

	// cluster stat check
	clusterFound, err := runtimeCache.GetCluster(plan.ClusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check cluster existence due to error: %s", err.Error()))
		return
	} else if clusterFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cluster %s not found.", plan.ClusterId))
		return
	} else if clusterFound.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster status is not in %s current status is %s", constants.ClusterStatusRunning, clusterFound.Status))
		return
	} else if clusterFound.ControlPlane.KubernetesVersion == plan.TargetVersion {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster already at version %s no need to upgrade.", plan.TargetVersion))
		return
	}

	if clusterFound.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to upgrade a mocked cluster '%s' ", clusterFound.ClusterId))
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	// prepare backup before upgrading
	if err = preBackup(planId, clusterFound, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("prepare backup faild before upgrading: %s", err.Error()))
	}

	if err := control_manager.ApplyUpgrade(plan, clusterFound, runByUser); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to upgrading cluster due to error: %s", err.Error()))
	}
}

func preBackup(planId string, cluster *schema.Cluster, runByUser string) error {
	cluster.Action = constants.ActionBackup
	cluster.Modified = time.Now().Format("2006-01-02T15:04:05Z07:00")
	cluster.Status = constants.ClusterStatusBackingUp

	backup := schema.Backup{
		Action:     backup_restore.Create,
		BackupName: planId,
	}

	return control_manager.BackupCreateWait(backup, cluster, runByUser)
}

func DeleteUpgradePlan(request *restful.Request, response *restful.Response) {
	planId := request.PathParameter("plan-id")
	if strings.TrimSpace(planId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "plan id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	if result, err := runtimeCache.GetUpgradePlan(planId); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get upgrade plan detail due to error: %s", err.Error()))
		return
	} else if result == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to find upgrade plan %s", planId))
		return
	} else if result.Status == constants.StatusProcessing {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Cannot delete the upgrade plan which currently executing")
		return
	} else if result.Status == constants.StatusError {
		utils.ResponseError(response, http.StatusPreconditionFailed, "Cannot delete the upgrade plan which in error state. It mays leads to control plane version not match issue which will block all upgrade function to this cluster. You should try continue upgrade operation or inform your administrator")
		return
	} else if result.Status == constants.StatusSuccessful {
		utils.ResponseError(response, http.StatusPreconditionFailed, "For audit issue you cannot delete the upgrade plan which in success state. ")
		return
	} else {
		if err := runtimeCache.DeleteUpgradePlan(planId); err != nil {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Failed to delete upgrade plan due to error: %s", err.Error()))
			return
		} else {
			response.WriteHeader(http.StatusOK)
		}
	}
}

func ListUpgradePlane(request *restful.Request, response *restful.Response) {
	clusterId := request.PathParameter("cluster-id")
	runtimeCache := cache.GetCurrentCache()

	if strings.TrimSpace(clusterId) == "" {
		if result, err := runtimeCache.GetUpgradePlanList(); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, "Internal server error !")
			return
		} else {
			if result == nil {
				response.WriteAsJson([]schema.UpgradableVersion{})
			} else {
				response.WriteAsJson(&result)
			}

		}
	} else {
		if result, err := runtimeCache.QueryUpgradePlan(func(plan schema.UpgradePlan) *schema.UpgradePlan {
			if plan.ClusterId == clusterId {
				return &plan
			} else {
				return nil
			}
		}); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, "Internal server error !")
			return
		} else {
			if result == nil {
				response.Write([]byte("{}"))
			} else {
				response.WriteAsJson(&result[0])
			}

		}
	}
}

func ListUpgradePlaneV2(request *restful.Request, response *restful.Response) {
	clusterId := request.QueryParameter("cluster-id")
	runtimeCache := cache.GetCurrentCache()

	if strings.TrimSpace(clusterId) == "" {
		if result, err := runtimeCache.GetUpgradePlanList(); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, "Internal server error !")
			return
		} else {
			if result == nil {
				response.WriteAsJson([]schema.UpgradableVersion{})
			} else {
				response.WriteAsJson(&result)
			}

		}
	} else {
		if result, err := runtimeCache.QueryUpgradePlan(func(plan schema.UpgradePlan) *schema.UpgradePlan {
			if plan.ClusterId == clusterId {
				return &plan
			} else {
				return nil
			}
		}); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, "Internal server error !")
			return
		} else {
			if result == nil {
				response.Write([]byte("{}"))
			} else {
				response.WriteAsJson(&result[0])
			}

		}
	}
}

func UpgradePlanDetail(request *restful.Request, response *restful.Response) {
	planId := request.PathParameter("plan-id")
	if strings.TrimSpace(planId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "plan id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	if result, err := runtimeCache.GetUpgradePlan(planId); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get upgrade plan detail due to error: %s", err.Error()))
		return
	} else if result == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find upgrade plan with id: %s", planId))
		return
	} else {
		response.WriteAsJson(&result)
	}
}

func ListUpgradableVersion(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	if result, err := runtimeCache.GetUpgradeVersionList(); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, "Internal server error !")
		return
	} else {
		if result == nil {
			response.WriteAsJson([]schema.UpgradableVersion{})
		} else {
			response.WriteAsJson(&result)
		}

	}
}

func UpdateUpgradableVersion(request *restful.Request, response *restful.Response) {
	upgrade := schema.UpgradableVersion{}
	if err := request.ReadEntity(&upgrade); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	versionName := request.PathParameter("version-name")
	if strings.TrimSpace(versionName) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "version name"))
		return
	}

	err := utils.ValidateUpgrade(upgrade)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Validation failed: %s", err.Error()))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	if found, err := runtimeCache.GetUpgradeVersion(versionName); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check upgrade version existence due to error: %s", upgrade.Name))
		return
	} else if found == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to found upgrade version %s, nothing to do !", versionName))
		return
	} else {
		runByUser := request.Attribute("run-by-user").(string)
		if strings.TrimSpace(runByUser) == "" {
			utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
			return
		}

		upgrade.LastModifiedBy = runByUser
		// version name cannot be changed
		// upgrade.Name = versionName
		upgrade.CreatedBy = found.CreatedBy
		upgrade.Created = found.Created
		now := time.Now().Format("2006-01-02T15:04:05Z07:00")
		upgrade.Modified = now
		// name cannot be change use original name
		upgrade.Name = versionName
		if err := runtimeCache.CreateOrUpdateUpgradeVersion(upgrade); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to update upgrade version to de due to error: %s", err.Error()))
			return
		} else {
			response.WriteAsJson(&upgrade)
		}
	}

}

func CreateUpgradableVersion(request *restful.Request, response *restful.Response) {
	upgrade := schema.UpgradableVersion{}
	if err := request.ReadEntity(&upgrade); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	err := utils.ValidateUpgrade(upgrade)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Validation failed: %s", err.Error()))
		return
	}

	runtimeCache := cache.GetCurrentCache()
	if found, err := runtimeCache.GetUpgradeVersion(upgrade.Name); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check upgrade version existence due to error: %s", upgrade.Name))
		return
	} else if found != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("UpgradeVersion %s already exists. Try update it.", upgrade.Name))
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	now := time.Now().Format("2006-01-02T15:04:05Z07:00")
	upgrade.Created = now
	upgrade.Modified = now
	upgrade.CreatedBy = runByUser
	upgrade.LastModifiedBy = runByUser
	// these property do not accept user input

	if err := runtimeCache.CreateOrUpdateUpgradeVersion(upgrade); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to save upgrade version to de due to error: %s", err.Error()))
		return
	} else {
		response.WriteAsJson(&upgrade)
	}
}

func DeleteUpgradableVersion(request *restful.Request, response *restful.Response) {
	versionName := request.PathParameter("version-name")
	if strings.TrimSpace(versionName) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "version name"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	if found, err := runtimeCache.GetUpgradeVersion(versionName); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check upgrade version existence due to error: %s", err.Error()))
		return
	} else if found == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to found upgrade version %s, nothing to do !", versionName))
		return
	}

	if err := runtimeCache.DeleteUpgradeVersion(versionName); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to delete upgrade version %s due to error: %s", versionName, err.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
}

func MakeUpgradePlan(request *restful.Request, response *restful.Response) {

	versionName := request.PathParameter("version-name")
	if strings.TrimSpace(versionName) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "version name"))
		return
	}
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "cluster id"))
		return
	}

	reg := regexp.MustCompile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)
	if !reg.MatchString(versionName) {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Upgrade version %s is not valid. Valid version format: [major].[sub].[min]. Example: 1.18.10", versionName))
		return
	}

	// check version can be found ?
	runtimeCache := cache.GetCurrentCache()
	foundVersion, err := runtimeCache.GetUpgradeVersion(versionName)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check upgrade version existence due to error: %s", err.Error()))
		return
	} else if foundVersion == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("UpgradeVersion %s not found.", versionName))
		return
	}

	// cluster stat check
	clusterFound, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check cluster existence due to error: %s", err.Error()))
		return
	} else if clusterFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cluster %s not found.", clusterId))
		return
	} else if clusterFound.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster status is not in %s current status is %s", constants.ClusterStatusRunning, clusterFound.Status))
		return
	} else if clusterFound.ControlPlane.KubernetesVersion == versionName {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster already at version %s no need to upgrade.", versionName))
		return
	}

	if clusterFound.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to upgrade a mocked cluster '%s' ", clusterFound.ClusterId))
		return
	}

	if clusterFound.ControlPlane.KubernetesVersion == "" {
		clusterFound.ControlPlane.KubernetesVersion = "1.18.6"
	}

	// ensure target version is greater than cluster version or we don`t have to do anything at all
	if result, err := util.NewerK8sVersion(clusterFound.ControlPlane.KubernetesVersion, versionName); err != nil {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to compare two version between current %s and target %s", clusterFound.ControlPlane.KubernetesVersion, versionName))
		return
	} else if result == clusterFound.ControlPlane.KubernetesVersion {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster current version %s is newer than target %s. Cannot downgrade to lower version.", clusterFound.ControlPlane.KubernetesVersion, versionName))
		return
	}

	if sameClusterPlan, err := runtimeCache.QueryUpgradePlan(func(plan schema.UpgradePlan) *schema.UpgradePlan {
		if plan.ClusterId == clusterFound.ClusterId {
			return &plan
		} else {
			return nil
		}
	}); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query check cluster %s upgrade plan existence due to error: %s", clusterFound.ClusterId, err.Error()))
		return
	} else {
		for _, plan := range sameClusterPlan {
			if plan.Status != constants.StatusSuccessful {
				utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster already have a upgrade plan with id %s . Remove or apply it before you can make new upgrade plan", sameClusterPlan[0].Id))
				return
			}
		}
	}

	// ensure no k8s installer operation is running
	if operations, err := runtimeCache.GetClusterOperation(*clusterFound); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, "Failed to check cluster operation stat")
		return
	} else {
		var runningOperationIDs []string
		for _, operation := range operations {
			if operation.Status == constants.StatusProcessing {
				runningOperationIDs = append(runningOperationIDs, operation.Id)
			}
		}
		if len(runningOperationIDs) > 0 {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("There were operations with id %s still processing?", strings.Join(runningOperationIDs, ",")))
			return
		}
	}

	plan, errPlan := PrintPlan(clusterFound, runtimeCache)
	if errPlan != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to make upgrade plan due to error: %s", errPlan.Error()))
		return
	}

	pkgCheckResult, isDepPackageReady, errCheckPkg := checkNodesPackage(*foundVersion, *clusterFound, runtimeCache)

	if errCheckPkg != nil {
		utils.ResponseError(response, http.StatusInternalServerError, errCheckPkg.Error())
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	_, errDownload := control_manager.UpgradeDownloadNewKubeadm(clusterFound, runByUser, versionName, dep.DepMap{
		"centos": {
			versionName: {
				"x86_64": {
					"kubeadm": "kubeadm",
				},
				"aarch64": {
					"kubeadm": "kubeadm",
				},
			},
		},
		"ubuntu": {
			versionName: {
				"x86_64": {
					"kubeadm": "kubeadm",
				},
				"aarch64": {
					"kubeadm": "kubeadm",
				},
			},
		},
	})

	if errDownload != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to install new kubeadm to node in order to get upgrade plan due to error: %s", errDownload.Error()))
		return
	}

	nodeStep := []schema.NodeStep{}
	task_breaker.CommonRunBatchWithSingleNode(clusterFound.Masters[0].NodeId, "", 30, [][]string{{"chmod", "+x", "/tmp/kubeadm"}, {"/tmp/kubeadm", "upgrade", "plan"}}, &nodeStep, true, false)
	operation := control_manager.CreateOperation("GetKubeadmUpgradePlan", clusterFound.ClusterId, "operation-"+uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeSingleTask
	operation.Step = map[int]*schema.Step{
		0: {
			Id:        "printKubeadmUpgradePlan",
			Name:      "Print upgrade plan",
			NodeSteps: nodeStep,
		},
	}

	printPlan, errPrintPlan := control_manager.DoSingleTask(&operation, clusterFound, map[string]string{})

	if errPrintPlan != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to print kubeadm upgrade plan due to error: %s", errPrintPlan.Error()))
		return
	}

	var reply schema.QueueReply

	if err := json.Unmarshal(printPlan, &reply); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed unmarshal data to reply object due to error: %s", err.Error()))
		return
	}

	corednsCurrentTag := ""
	corednsUpgradeToTag := ""
	etcdCurrentTag := ""
	etcdUpgradeToTag := ""

	rawResult := strings.Split(reply.ReturnData["1"], "\n")

	for _, line := range rawResult {
		if strings.HasPrefix(strings.ToLower(line), "coredns") {
			corednsSet := strings.Split(line, " ")
			corednsCurrentTag, corednsUpgradeToTag = getVersionFromKubeadmPlan(corednsSet)
		}
		if strings.HasPrefix(strings.ToLower(line), "etcd") {
			etcdSet := strings.Split(line, " ")
			etcdCurrentTag, etcdUpgradeToTag = getVersionFromKubeadmPlan(etcdSet)
		}
	}

	currentClusterVersion := "v" + clusterFound.ControlPlane.KubernetesVersion
	k8sComUpgradeToVersion := "v" + versionName

	var imageCheckList = []schema.ImageCheckState{
		{
			Name:         "kube-apiserver",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-controller-manager",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-proxy",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-scheduler",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "coredns",
			CurrentTag:   corednsCurrentTag,
			UpgradeToTag: corednsUpgradeToTag,
			Status:       "notReady",
		},
		{
			Name:         "etcd",
			CurrentTag:   etcdCurrentTag,
			UpgradeToTag: etcdUpgradeToTag,
			Status:       "notReady",
		},
	}

	isImageReady := false

	registryAddress := ""

	if clusterFound.ContainerRuntime.PrivateRegistryAddress != "" {
		port := 4000
		if clusterFound.ContainerRuntime.PrivateRegistryPort != 0 {
			port = clusterFound.ContainerRuntime.PrivateRegistryPort
		}
		registryAddress = fmt.Sprintf("http://%s:%s", clusterFound.ContainerRuntime.PrivateRegistryAddress, strconv.Itoa(port))
		privateRegistryImages, errListImg := docker.ListImagesV2(registryAddress)
		if errListImg == nil {
			log.Debug("Do image check")
			isImageReady = checkPrivateRegistryImages(imageCheckList, privateRegistryImages)
		} else {
			log.Errorf("Failed to list private registry %s due to error %s.", registryAddress, errListImg.Error())
		}
	} else {
		log.Warn("Private registry is not set, skip image check and set image check stat to not ready")
	}

	planResultMessage := ""
	if !isDepPackageReady {
		planResultMessage = "There are package dependencies issue in cluster nodes. Please look at package check section to see what`s not ready\n"
	}

	if !isImageReady {
		planResultMessage += fmt.Sprintf("There are images not found in cluster private registry %s or it`s a online registry?. Please look at image check section to see what`s not ready\n", registryAddress)
	}

	if len(planResultMessage) == 0 {
		planResultMessage = "Everything looking good upgrade plan made and ready to be trigger by user"
	}

	if len(clusterFound.Masters) == 1 {
		planResultMessage += "Cluster only have one master any service rely on api-server will temporary unavailable"
	}

	upgradePlan := schema.UpgradePlan{
		ClusterId:         clusterId,
		ImageChangePlan:   imageCheckList,
		IsPackageReady:    isDepPackageReady,
		NodeChangePlan:    pkgCheckResult,
		IsImageReady:      isImageReady,
		Plan:              plan,
		TargetVersion:     versionName,
		MessagePlanResult: planResultMessage,
		Status:            "ReadyToUpgrade",
	}

	if isDepPackageReady && isImageReady {
		upgradePlan.Id = "upgrade-plan-" + uuid.New().String()
		if err := runtimeCache.CreateOrUpdateUpgradePlan(upgradePlan); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed save upgrade plan to db due to error: %s", err.Error()))
			return
		}
	}

	if !isImageReady || !isDepPackageReady {
		response.WriteHeader(http.StatusPreconditionFailed)
	}

	response.WriteAsJson(&upgradePlan)
}

func MakeUpgradePlanV2(request *restful.Request, response *restful.Response) {
	versionName := request.PathParameter("version-name")
	if strings.TrimSpace(versionName) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find parameter: %s", "version name"))
		return
	}
	clusterId := request.PathParameter("cluster-id")
	if strings.TrimSpace(clusterId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find parameter: %s", "cluster id"))
		return
	}

	reg := regexp.MustCompile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)
	if !reg.MatchString(versionName) {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Upgrade version %s is not valid. Valid version format: [major].[sub].[min]. Example: 1.18.10", versionName))
		return
	}

	// check version can be found ?
	runtimeCache := cache.GetCurrentCache()
	foundVersion, err := runtimeCache.GetUpgradeVersion(versionName)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check upgrade version existence due to error: %s", err.Error()))
		return
	} else if foundVersion == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("UpgradeVersion %s not found.", versionName))
		return
	}

	// cluster stat check
	clusterFound, err := runtimeCache.GetCluster(clusterId)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check cluster existence due to error: %s", err.Error()))
		return
	} else if clusterFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cluster %s not found.", clusterId))
		return
	} else if clusterFound.Status != constants.ClusterStatusRunning {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster status is not in %s current status is %s", constants.ClusterStatusRunning, clusterFound.Status))
		return
	} else if clusterFound.ControlPlane.KubernetesVersion == versionName {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster already at version %s no need to upgrade.", versionName))
		return
	}

	if clusterFound.Mock {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to upgrade a mocked cluster '%s' ", clusterFound.ClusterId))
		return
	}

	if clusterFound.ControlPlane.KubernetesVersion == "" {
		clusterFound.ControlPlane.KubernetesVersion = "1.18.6"
	}

	// ensure target version is greater than cluster version or we don`t have to do anything at all
	if result, err := util.NewerK8sVersion(clusterFound.ControlPlane.KubernetesVersion, versionName); err != nil {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Unable to compare two version between current %s and target %s", clusterFound.ControlPlane.KubernetesVersion, versionName))
		return
	} else if result == clusterFound.ControlPlane.KubernetesVersion {
		utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster current version %s is newer than target %s. Cannot downgrade to lower version.", clusterFound.ControlPlane.KubernetesVersion, versionName))
		return
	}

	if sameClusterPlan, err := runtimeCache.QueryUpgradePlan(func(plan schema.UpgradePlan) *schema.UpgradePlan {
		if plan.ClusterId == clusterFound.ClusterId {
			return &plan
		} else {
			return nil
		}
	}); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to query check cluster %s upgrade plan existence due to error: %s", clusterFound.ClusterId, err.Error()))
		return
	} else {
		for _, plan := range sameClusterPlan {
			if plan.Status != constants.StatusSuccessful {
				utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Cluster already have a upgrade plan with id %s . Remove or apply it before you can make new upgrade plan", sameClusterPlan[0].Id))
				return
			}
		}
	}

	// ensure no k8s installer operation is running
	if operations, err := runtimeCache.GetClusterOperation(*clusterFound); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, "Failed to check cluster operation stat")
		return
	} else {
		var runningOperationIDs []string
		for _, operation := range operations {
			if operation.Status == constants.StatusProcessing {
				runningOperationIDs = append(runningOperationIDs, operation.Id)
			}
		}
		if len(runningOperationIDs) > 0 {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("There were operations with id %s still processing?", strings.Join(runningOperationIDs, ",")))
			return
		}
	}

	plan, errPlan := PrintPlan(clusterFound, runtimeCache)
	if errPlan != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to make upgrade plan due to error: %s", errPlan.Error()))
		return
	}

	pkgCheckResult, isDepPackageReady, errCheckPkg := checkNodesPackage(*foundVersion, *clusterFound, runtimeCache)

	if errCheckPkg != nil {
		utils.ResponseError(response, http.StatusInternalServerError, errCheckPkg.Error())
		return
	}

	runByUser := request.Attribute("run-by-user").(string)
	if strings.TrimSpace(runByUser) == "" {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to find run by user!!!")
		return
	}

	_, errDownload := control_manager.UpgradeDownloadNewKubeadm(clusterFound, runByUser, versionName, dep.DepMap{
		"centos": {
			versionName: {
				"x86_64": {
					"kubeadm": "kubeadm",
				},
				"aarch64": {
					"kubeadm": "kubeadm",
				},
			},
		},
		"ubuntu": {
			versionName: {
				"x86_64": {
					"kubeadm": "kubeadm",
				},
				"aarch64": {
					"kubeadm": "kubeadm",
				},
			},
		},
	})

	if errDownload != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to install new kubeadm to node in order to get upgrade plan due to error: %s", errDownload.Error()))
		return
	}

	nodeStep := []schema.NodeStep{}
	task_breaker.CommonRunBatchWithSingleNode(clusterFound.Masters[0].NodeId, "", 30, [][]string{{"chmod", "+x", "/tmp/kubeadm"}, {"/tmp/kubeadm", "upgrade", "plan"}}, &nodeStep, true, false)
	operation := control_manager.CreateOperation("GetKubeadmUpgradePlan", clusterFound.ClusterId, "operation-"+uuid.New().String(), constants.StatusProcessing)
	operation.OperationType = constants.OperationTypeSingleTask
	operation.Step = map[int]*schema.Step{
		0: {
			Id:        "printKubeadmUpgradePlan",
			Name:      "Print upgrade plan",
			NodeSteps: nodeStep,
		},
	}

	printPlan, errPrintPlan := control_manager.DoSingleTask(&operation, clusterFound, map[string]string{})

	if errPrintPlan != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to print kubeadm upgrade plan due to error: %s", errPrintPlan.Error()))
		return
	}

	var reply schema.QueueReply

	if err := json.Unmarshal(printPlan, &reply); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed unmarshal data to reply object due to error: %s", err.Error()))
		return
	}

	corednsCurrentTag := ""
	corednsUpgradeToTag := ""
	etcdCurrentTag := ""
	etcdUpgradeToTag := ""

	rawResult := strings.Split(reply.ReturnData["1"], "\n")

	for _, line := range rawResult {
		if strings.HasPrefix(strings.ToLower(line), "coredns") {
			corednsSet := strings.Split(line, " ")
			corednsCurrentTag, corednsUpgradeToTag = getVersionFromKubeadmPlan(corednsSet)
		}
		if strings.HasPrefix(strings.ToLower(line), "etcd") {
			etcdSet := strings.Split(line, " ")
			etcdCurrentTag, etcdUpgradeToTag = getVersionFromKubeadmPlan(etcdSet)
		}
	}

	currentClusterVersion := "v" + clusterFound.ControlPlane.KubernetesVersion
	k8sComUpgradeToVersion := "v" + versionName

	var imageCheckList = []schema.ImageCheckState{
		{
			Name:         "kube-apiserver",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-controller-manager",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-proxy",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "kube-scheduler",
			CurrentTag:   currentClusterVersion,
			UpgradeToTag: k8sComUpgradeToVersion,
			Status:       "notReady",
		},
		{
			Name:         "coredns",
			CurrentTag:   corednsCurrentTag,
			UpgradeToTag: corednsUpgradeToTag,
			Status:       "notReady",
		},
		{
			Name:         "etcd",
			CurrentTag:   etcdCurrentTag,
			UpgradeToTag: etcdUpgradeToTag,
			Status:       "notReady",
		},
	}

	isImageReady := false

	registryAddress := ""

	if clusterFound.ContainerRuntime.PrivateRegistryAddress != "" {
		port := 4000
		if clusterFound.ContainerRuntime.PrivateRegistryPort != 0 {
			port = clusterFound.ContainerRuntime.PrivateRegistryPort
		}
		registryAddress = fmt.Sprintf("http://%s:%s", clusterFound.ContainerRuntime.PrivateRegistryAddress, strconv.Itoa(port))
		privateRegistryImages, errListImg := docker.ListImagesV2(registryAddress)
		if errListImg == nil {
			log.Debug("Do image check")
			isImageReady = checkPrivateRegistryImages(imageCheckList, privateRegistryImages)
		} else {
			log.Errorf("Failed to list private registry %s due to error %s.", registryAddress, errListImg.Error())
		}
	} else {
		log.Warn("Private registry is not set, skip image check and set image check stat to not ready")
	}

	planResultMessage := ""
	if !isDepPackageReady {
		planResultMessage = "There are package dependencies issue in cluster nodes. Please look at package check section to see what`s not ready\n"
	}

	if !isImageReady {
		planResultMessage += fmt.Sprintf("There are images not found in cluster private registry %s or it`s a online registry?. Please look at image check section to see what`s not ready\n", registryAddress)
	}

	if len(planResultMessage) == 0 {
		planResultMessage = "Everything looking good upgrade plan made and ready to be trigger by user"
	}

	if len(clusterFound.Masters) == 1 {
		planResultMessage += "Cluster only have one master any service rely on api-server will temporary unavailable"
	}

	upgradePlan := schema.UpgradePlan{
		ClusterId:         clusterId,
		ImageChangePlan:   imageCheckList,
		IsPackageReady:    isDepPackageReady,
		NodeChangePlan:    pkgCheckResult,
		IsImageReady:      isImageReady,
		Plan:              plan,
		TargetVersion:     versionName,
		MessagePlanResult: planResultMessage,
		Status:            "ReadyToUpgrade",
	}

	if isDepPackageReady && isImageReady {
		upgradePlan.Id = "upgrade-plan-" + uuid.New().String()
		if err := runtimeCache.CreateOrUpdateUpgradePlan(upgradePlan); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed save upgrade plan to db due to error: %s", err.Error()))
			return
		}
	}

	if !isImageReady || !isDepPackageReady {
		response.WriteHeader(http.StatusPreconditionFailed)
	}

	response.WriteAsJson(&upgradePlan)
}

func PrintPlan(cluster *schema.Cluster, runtimeCache cache.ICache) ([]string, error) {
	var result []string
	masterMap := map[string]byte{}
	for _, master := range cluster.Masters {
		if node, err := runtimeCache.GetNodeInformation(master.NodeId); err != nil {
			return nil, err
		} else {
			masterMap[node.Id] = 0
			result = append(result, fmt.Sprintf("Cordon master %s", node.Ipv4DefaultIp))
			result = append(result, fmt.Sprintf("Upgrade package for master %s", node.Ipv4DefaultIp))
			result = append(result, fmt.Sprintf("Upgrade kubernetes images for master %s", node.Ipv4DefaultIp))
			result = append(result, fmt.Sprintf("UnCordon master %s", node.Ipv4DefaultIp))
		}
	}
	for _, worker := range cluster.Workers {
		if _, exists := masterMap[worker.NodeId]; !exists {
			if node, err := runtimeCache.GetNodeInformation(worker.NodeId); err != nil {
				return nil, err
			} else {

				result = append(result, fmt.Sprintf("Cordon woker %s", node.Ipv4DefaultIp))
				result = append(result, fmt.Sprintf("Upgrade package for worker %s", node.Ipv4DefaultIp))
				result = append(result, fmt.Sprintf("Upgrade kubernetes images for worker %s", node.Ipv4DefaultIp))
				result = append(result, fmt.Sprintf("UnCordon master %s", node.Ipv4DefaultIp))
				result = append(result, fmt.Sprintf("Restart kubelet for worker %s", node.Ipv4DefaultIp))
			}
		}
	}
	for _, master := range cluster.Masters {
		if node, err := runtimeCache.GetNodeInformation(master.NodeId); err != nil {
			return nil, err
		} else {
			result = append(result, fmt.Sprintf("Restart kubelet for master %s", node.Ipv4DefaultIp))
		}
	}

	return result, nil
}

func checkNodesPackage(versionSet schema.UpgradableVersion, cluster schema.Cluster, runtimeCache cache.ICache) ([]schema.UpgradeNodeCheckState, bool, error) {
	var results []schema.UpgradeNodeCheckState
	isAllNodesReady := true
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	ip := network.GetDefaultIP(true)
	resourceServerUrl := fmt.Sprintf("http://%s:%d", ip.String(), config.ApiServer.ResourceServerPort)

	doneCheckedNodes, checkedMasterUpgradeState, isAllMasterReady, errMaster := checkNodeLayout(resourceServerUrl, versionSet, nil, cluster.Masters, runtimeCache)
	if !isAllMasterReady {
		isAllNodesReady = false
	}
	if errMaster != nil {
		return nil, false, errMaster
	} else {
		results = append(results, checkedMasterUpgradeState...)
	}

	_, checkedWorkerUpgradeState, isAllWorkerReady, errWorker := checkNodeLayout(resourceServerUrl, versionSet, doneCheckedNodes, cluster.Workers, runtimeCache)
	if !isAllWorkerReady {
		isAllNodesReady = false
	}

	if errWorker != nil {
		return nil, false, errWorker
	} else {
		results = append(results, checkedWorkerUpgradeState...)
	}

	return results, isAllNodesReady, nil
}

func checkNodeLayout(resourceServerUrl string, versionSet schema.UpgradableVersion, doneCheckedList map[string]byte, nodes []schema.ClusterNode, runtimeCache cache.ICache) (map[string]byte, []schema.UpgradeNodeCheckState, bool, error) {
	if len(nodes) == 0 {
		return nil, nil, true, nil
	}
	isAllReady := true
	doneCheckedNodes := map[string]byte{}
	var results []schema.UpgradeNodeCheckState
	for _, node := range nodes {
		if _, exists := doneCheckedList[node.NodeId]; exists {
			continue
		}
		doneCheckedNodes[node.NodeId] = 0
		if nodeFound, err := runtimeCache.GetNodeInformation(node.NodeId); err != nil {
			return doneCheckedNodes, nil, false, errors.New(fmt.Sprintf("Failed to find node %s information due to error: %s ", node.NodeId, err.Error()))
		} else if nodeFound == nil {
			return doneCheckedNodes, nil, false, errors.New(fmt.Sprintf("Cannot check node %s package dependencies due to unable to find node %s information ", node.NodeId, node.NodeId))
		} else {
			result, isNodeReady := checkNodePackage(versionSet, *nodeFound, resourceServerUrl)
			if !isNodeReady {
				isAllReady = false
			}
			results = append(results, result)
		}
	}
	return doneCheckedNodes, results, isAllReady, nil
}

func checkNodePackage(versionSet schema.UpgradableVersion, node schema.NodeInformation, resourceServerAddress string) (schema.UpgradeNodeCheckState, bool) {
	isAllNodesReady := false
	cpuArch := node.SystemInfo.Kernel.Architecture
	nodeCheckState := schema.UpgradeNodeCheckState{
		NodeID:   node.Id,
		OSFamily: node.SystemInfo.OS.Vendor,
		CpuArch:  cpuArch,
		NodeIPV4: node.Ipv4DefaultIp,
	}
	var depPackages []schema.OSPackage
	switch node.SystemInfo.OS.Vendor {
	case constants.OSFamilyCentos:
		if versionSet.Centos == nil {
			return schema.UpgradeNodeCheckState{}, true
		}
		if cpuArch == constants.CpuArchX86 {
			depPackages = versionSet.Centos.OSVersions[node.SystemInfo.OS.Version].X8664.RPMS
		} else if cpuArch == constants.CpuAarch64 {
			depPackages = versionSet.Centos.OSVersions[node.SystemInfo.OS.Version].Aarch64.RPMS
		}
	case constants.OSFamilyUbuntu:
		if versionSet.Centos == nil {
			return schema.UpgradeNodeCheckState{}, true
		}
		if cpuArch == constants.CpuArchX86 {
			depPackages = versionSet.Ubuntu.OSVersions[node.SystemInfo.OS.Version].X8664.Debs
		} else if cpuArch == constants.CpuAarch64 {
			depPackages = versionSet.Ubuntu.OSVersions[node.SystemInfo.OS.Version].Aarch64.Debs
		}
	}

	var result []schema.PackageCheckState
	result, isAllNodesReady = checkUpgradeDepPkgWithOSFamily(depPackages, resourceServerAddress, node.SystemInfo.OS.Vendor, node.SystemInfo.OS.Version, versionSet.Name, cpuArch)
	nodeCheckState.PackageCheckResult = result
	return nodeCheckState, isAllNodesReady
}

func checkUpgradeDepPkgWithOSFamily(depPackages []schema.OSPackage, resourceServerAddress, osFamily, osVersion, k8sVersion, cpuArch string) ([]schema.PackageCheckState, bool) {
	isNodeAllPkgReady := true
	var pkgCheckResults []schema.PackageCheckState
	for _, pkg := range depPackages {
		url := fmt.Sprintf("%s/%s/%s/%s/%s/package/%s", resourceServerAddress, k8sVersion, osFamily, osVersion, cpuArch, pkg.Name)
		if util.CheckHttpExists(url) {
			pkgCheckResults = append(pkgCheckResults, schema.PackageCheckState{
				Name:   pkg.Name,
				Status: "Ready",
			})
		} else {
			isNodeAllPkgReady = false
			pkgCheckResults = append(pkgCheckResults, schema.PackageCheckState{
				Name:   pkg.Name,
				Status: "NotReady",
			})
		}
	}
	return pkgCheckResults, isNodeAllPkgReady
}

func getVersionFromKubeadmPlan(source []string) (string, string) {
	var oldVersion, newVersion string
	found := false
	index := len(source) - 1
	for index > 0 {
		if source[index] != "" {
			if !found {
				newVersion = source[index]
				found = true
			} else {
				oldVersion = source[index]
				break
			}
		}
		index--
	}
	return oldVersion, newVersion
}

func checkPrivateRegistryImages(checkList []schema.ImageCheckState, privateRegistryImages map[string]docker.PRV2Image) bool {
	isAllImageReady := true
	for index, image := range checkList {
		breakOutLoop := false
		if _, exists := privateRegistryImages[image.Name]; exists {
			for _, tag := range privateRegistryImages[image.Name].Tags {
				if tag == image.UpgradeToTag {
					checkList[index].Status = "Ready"
					breakOutLoop = true
					break
				}
			}
		}
		if breakOutLoop {
			continue
		}
		isAllImageReady = false
	}
	return isAllImageReady
}
