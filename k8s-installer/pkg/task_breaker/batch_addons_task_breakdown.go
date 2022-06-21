package task_breaker

import (
	"reflect"

	"k8s-installer/schema/plugable"

	serverConfig "k8s-installer/pkg/config/server"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type BatchAddonsTaskBreakDown struct {
	Operation          *schema.Operation
	Cluster            *schema.Cluster
	OriginalCluster    schema.Cluster
	NewAddons          schema.Addons
	FirstMasterID      string
	Config             serverConfig.Config
	NodeCollectionList schema.NodeInformationCollection
	SumLicenseLabel    uint16
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config serverConfig.Config) error {
	// batch addons will never involve with cluster creation or destroying
	return nil
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) BreakDownTask() (schema.Operation, error) {
	// check whether cluster lb change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createClusterLBDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.ClusterLB), batchAddonsTaskBreakDown.NewAddons.ClusterLB.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.ClusterLB.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.ClusterLB.Enable = batchAddonsTaskBreakDown.NewAddons.ClusterLB.Enable
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.ClusterLB = batchAddonsTaskBreakDown.OriginalCluster.ClusterLB
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether openstack cloud provider change stat
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createCloudProviderDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.CloudProvider.OpenStack), batchAddonsTaskBreakDown.NewAddons.CloudProvider.OpenStack.Enable,
		func(fromEnableToDisable bool) {
			// stat change so we need to replace the stat and save to db later
			// if stat changed from enable to disable we simply save the stat only and keep the original property
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.CloudProvider.OpenStack.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.CloudProvider.OpenStack = batchAddonsTaskBreakDown.NewAddons.CloudProvider.OpenStack
			}
		}, func() {
			// didn't change enable stat , keep original stat
			batchAddonsTaskBreakDown.Cluster.CloudProvider.OpenStack = batchAddonsTaskBreakDown.OriginalCluster.CloudProvider.OpenStack
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether efk change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createEFKDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.EFK), batchAddonsTaskBreakDown.NewAddons.EFK.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.EFK.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.EFK = batchAddonsTaskBreakDown.NewAddons.EFK
			}
		}, func() {
			// didn't change enable stat , keep original stat
			batchAddonsTaskBreakDown.Cluster.EFK = batchAddonsTaskBreakDown.OriginalCluster.EFK
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether postgres operator change state
	if err := batchAddonsTaskBreakDown.breakDownTask(
		batchAddonsTaskBreakDown.createPostgresOperatorDeploy(),
		checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.PostgresOperator),
		batchAddonsTaskBreakDown.NewAddons.PostgresOperator.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.PostgresOperator.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.PostgresOperator = batchAddonsTaskBreakDown.NewAddons.PostgresOperator
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.PostgresOperator = batchAddonsTaskBreakDown.OriginalCluster.PostgresOperator
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether middle platform change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createMiddlePlatformDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.MiddlePlatform), batchAddonsTaskBreakDown.NewAddons.MiddlePlatform.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.MiddlePlatform.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.MiddlePlatform = batchAddonsTaskBreakDown.NewAddons.MiddlePlatform
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.MiddlePlatform = batchAddonsTaskBreakDown.OriginalCluster.MiddlePlatform
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether console change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createConsoleDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.Console), batchAddonsTaskBreakDown.NewAddons.Console.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.Console.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.Console = batchAddonsTaskBreakDown.NewAddons.Console
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.Console = batchAddonsTaskBreakDown.OriginalCluster.Console
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether helm change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createHelmDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.Helm), batchAddonsTaskBreakDown.NewAddons.Helm.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.Helm.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.Helm = batchAddonsTaskBreakDown.NewAddons.Helm
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.Helm = batchAddonsTaskBreakDown.OriginalCluster.Helm
		}); err != nil {
		return schema.Operation{}, err
	}

	// check whether GAP change state
	if err := batchAddonsTaskBreakDown.breakDownTask(batchAddonsTaskBreakDown.createGAPDeploy(), checkAddonsStructIsNil(batchAddonsTaskBreakDown.Cluster.GAP), batchAddonsTaskBreakDown.NewAddons.GAP.Enable,
		func(fromEnableToDisable bool) {
			if fromEnableToDisable {
				batchAddonsTaskBreakDown.Cluster.GAP.Enable = false
			} else {
				batchAddonsTaskBreakDown.Cluster.GAP = batchAddonsTaskBreakDown.NewAddons.GAP
			}
		}, func() {
			batchAddonsTaskBreakDown.Cluster.GAP = batchAddonsTaskBreakDown.OriginalCluster.GAP
		}); err != nil {
		return schema.Operation{}, err
	}

	return *batchAddonsTaskBreakDown.Operation, nil
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) breakDownTask(addons IAddOns, originalStat, changeStat bool, changeStatHandler func(fromEnableToDisable bool), statKeep func()) error {
	if originalStat != changeStat {
		log.Debugf("Detect addons %s stat change from enable:%t to enable:%t", addons.GetAddOnName(), originalStat, changeStat)
		changeStatHandler(originalStat)
		if err := deployOrRemoveAddons(addons, batchAddonsTaskBreakDown.Operation, *batchAddonsTaskBreakDown.Cluster, batchAddonsTaskBreakDown.Config); err != nil {
			return err
		}
	} else {
		// keep the old stat
		// if user keep the addons enable switch on but change one of it`s property, in that case wo do not reinstall that addons and we should keep the original data
		statKeep()
	}
	return nil
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createClusterLBDeploy() AddOnClusterLB {
	return AddOnClusterLB{
		Name:        "ClusterLBAddOn",
		TaskTimeOut: 5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createEFKDeploy() AddOnEFK {

	isIngressEnable := false
	isConsoleEneble := false
	var consoleNamespace string
	if batchAddonsTaskBreakDown.Cluster.Console != nil {
		isConsoleEneble = batchAddonsTaskBreakDown.Cluster.Console.Enable
		consoleNamespace = batchAddonsTaskBreakDown.Cluster.Console.Namespace
	}

	if batchAddonsTaskBreakDown.Cluster.Ingress != nil {
		isIngressEnable = batchAddonsTaskBreakDown.Cluster.Console.Enable
	}
	return AddOnEFK{
		Name:             "EFKAddOn",
		FirstMasterId:    batchAddonsTaskBreakDown.FirstMasterID,
		TaskTimeOut:      5,
		IsIngressEnable:  isIngressEnable,
		IsConsoleEneble:  isConsoleEneble,
		ConsoleNamespace: consoleNamespace,
	}.SetDataWithPlugin(batchAddonsTaskBreakDown.Cluster.EFK, batchAddonsTaskBreakDown.SumLicenseLabel, *batchAddonsTaskBreakDown.Cluster).(AddOnEFK)
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createCloudProviderDeploy() AddOnOpenStackCloudProvider {
	return AddOnOpenStackCloudProvider{
		FullNodeList: createFullNodeList(*batchAddonsTaskBreakDown.Cluster, batchAddonsTaskBreakDown.NodeCollectionList),
		Name:         "OpenStackCloudProviderAddOn",
		TaskTimeOut:  5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createPostgresOperatorDeploy() AddOnPostgresOperator {
	return AddOnPostgresOperator{
		Name:        "PostgresOperatorAddOn",
		TaskTimeOut: 5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createMiddlePlatformDeploy() AddOnMiddlePlatform {
	return AddOnMiddlePlatform{
		Name:        "MiddlePlatformAddOn",
		TaskTimeOut: 5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createConsoleDeploy() AddOnConsole {
	return AddOnConsole{
		Name:        "Console",
		TaskTimeOut: 5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createHelmDeploy() AddOnHelm {
	return AddOnHelm{
		Name:        "Helm",
		TaskTimeOut: 5,
	}
}

func (batchAddonsTaskBreakDown *BatchAddonsTaskBreakDown) createGAPDeploy() AddOnGAP {
	return AddOnGAP{
		Name:          "GAPAddOn",
		TaskTimeOut:   5,
		FirstMasterId: batchAddonsTaskBreakDown.Cluster.Masters[0].NodeId,
	}
}

func checkAddonsStructIsNil(addonsDBStruct plugable.IPlugAble) bool {
	if addonsDBStruct == nil || reflect.ValueOf(addonsDBStruct).IsNil() {
		return false
	} else {
		return addonsDBStruct.IsEnable()
	}
}
