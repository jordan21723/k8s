package cache

import (
	"k8s-installer/pkg/config/cache"
	"k8s-installer/pkg/config/client"
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/coredns"
	"k8s-installer/schema"
)

/*
cache layer handler all data
you don`t need to worry about how data is handled by program
cache layer ensure all data is well handled
*/

type ICache interface {
	IUserRole
	IOperation
	IClusterNode
	IUpgrade
	IRegion
	IDNS
	SyncFromDatabase() error
	initialization(config etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) error
	SaveOrUpdateServerRuntimeConfig(serverId string, config server.Config) error
	GetServerRuntimeConfig(serverId string) server.Config
	SaveOrUpdateClientRuntimeConfig(clientId string, config client.Config) error
	GetClientRuntimeConfig(clientId string) client.Config
}

type IClusterNode interface {
	GetClusterCollection() (schema.ClusterCollection, error)
	GetCluster(clusterId string) (*schema.Cluster, error)
	GetClusterWithShortId(shortClusterId string) (*schema.Cluster, error)
	SaveOrUpdateClusterCollection(clusterId string, cluster schema.Cluster) error
	DeleteClusterCollection(clusterId string) error
	GetNodeInformationCollection() (schema.NodeInformationCollection, error)
	GetNodeInformationCollectionWithPage(pageIndex, pageSize int64) (schema.NodeInformationCollection, int64, error)
	GetNodeInformationWithCluster(clusterId string) ([]schema.NodeInformation, error)
	GetNodeInformation(nodeId string) (*schema.NodeInformation, error)
	SaveOrUpdateNodeInformation(nodeId string, nodeInformation schema.NodeInformation) error
	DeleteNodeInformation(nodeId string) error
	SetNodeStatus(nodeId, status string) error
	AddClusterNodeRelationShip(clusterId string, NodeId []string) error
	RemoveClusterNodeRelationShip(clusterId string, NodeId []string) error
	RemoveAllClusterNodeRelationShip(clusterId string) error
	GetClusterNodeRelationShip(clusterId string) (map[string]byte, error)
	GetAvailableNodesMap(pageIndex, pageSize int64, filter func(information schema.NodeInformation) *schema.NodeInformation) (schema.NodeInformationCollection, int64, error)
	CreateOrUpdadeLicense(license schema.LicenseInfo) error
	GetLicense() (*schema.LicenseInfo, error)
	SaveClusterTemplate(*schema.ClusterTemplate) error
	GetTemplate() (*schema.ClusterTemplate, error)
}

type IUserRole interface {
	CreateOrUpdateUser(userId string, user schema.User) error
	GetUserList() (schema.UserCollection, error)
	GetUser(username string) (*schema.User, error)
	Login(username, password string) (*schema.User, error)
	DeleteUser(userId string)
	CreateOrUpdateRole(roleId string, role schema.Role) error
	GetRoleList() (schema.RoleCollection, error)
	GetRole(roleId string) (*schema.Role, error)
	DeleteRole(roleId string)
}

type IOperation interface {
	GetOperationCollection() (schema.OperationCollection, error)
	GetOperation(operationId string) (*schema.Operation, error)
	GetClusterOperation(cluster schema.Cluster) ([]schema.Operation, error)
	GetClusterOperationDoNotLoadStep(cluster schema.Cluster) ([]schema.Operation, error)
	SaveOrUpdateOperationCollection(operationId string, operation schema.Operation) error
	SetOperationStatus(operationId, status string) error
	DeleteOperation(operationId string) error
}

type IUpgrade interface {
	GetUpgradeVersion(versionName string) (*schema.UpgradableVersion, error)
	GetUpgradeVersionList() ([]schema.UpgradableVersion, error)
	GetUpgradeVersionCollection() (schema.UpgradeVersionCollection, error)
	CreateOrUpdateUpgradeVersion(upgradeVersion schema.UpgradableVersion) error
	DeleteUpgradeVersion(version string) error
	CreateOrUpdateUpgradePlan(plan schema.UpgradePlan) error
	GetUpgradePlanCollection() (schema.UpgradePlanCollection, error)
	GetUpgradePlanList() ([]schema.UpgradePlan, error)
	QueryUpgradePlan(filter func(plan schema.UpgradePlan) *schema.UpgradePlan) ([]schema.UpgradePlan, error)
	GetUpgradePlan(id string) (*schema.UpgradePlan, error)
	DeleteUpgradePlan(id string) error
}

type IRegion interface {
	CreateOrUpdateRegion(region schema.Region) error
	GetRegionList() (schema.RegionCollection, error)
	GetRegion(regionID string) (*schema.Region, error)
	DeleteRegion(regionID string) error
}

type IDNS interface {
	CreateOrUpdateTopLevelDomain(domain coredns.TopLevelDomain) error
	GetTopLevelDomainList() (schema.TopLevelDomainCollection, error)
	GetTopLevelDomain(domain string) (*coredns.TopLevelDomain, error)
	DeleteTopLevelDomain(domain string) error

	CreateOrUpdateSubDomain(domain coredns.DNSDomain) error
	GetSubDomainList(topLevelDomain string) (schema.SubDomainCollection, error)
	GetAllSubDomainList() (schema.SubDomainCollection, error)
	GetSubDomain(topLevelDomain, domain string) (*coredns.DNSDomain, error)
	GetSubDomainOfTls(topLevelDomain string) (schema.SubDomainCollection, error)
	DeleteSubDomain(topLevelDomain, domain string) error
}

var localRam *LocalRam
var noCache *NoCache
var currentCache ICache

func initOrGetLocalRam(config etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) *LocalRam {
	if localRam == nil {
		localRam = &LocalRam{}
		localRam.initialization(config, operationLoader)
	}
	currentCache = localRam
	return localRam
}

func initOrGetNoCache(config etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) *NoCache {
	if noCache == nil {
		noCache = &NoCache{}
		noCache.initialization(config, operationLoader)
	}
	currentCache = noCache
	return noCache
}

func InitCacheRuntime(config cache.CacheConfig, etcdConfig etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) ICache {
	switch config.CacheRuntime {
	case "local-ram":
		return initOrGetLocalRam(etcdConfig, operationLoader)
	case "no-cache":
		return initOrGetNoCache(etcdConfig, operationLoader)
	default:
		return initOrGetLocalRam(etcdConfig, operationLoader)
	}
}

func GetCurrentCache() ICache {
	return currentCache
}
