package cache

import (
	"k8s-installer/pkg/config/client"
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/schema"
)

/*
local cache should only use with single server
*/
type LocalRam struct {
	UserRoleCache
	OperationCache
	ClusterNodeCache
	UpgradeCache
	RegionCache
	IDNS
}

func (g *LocalRam) initialization(etcdConfig etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) error {
	g.UserRoleCache = UserRoleCache{etcdConfig: etcdConfig}
	g.OperationCache = OperationCache{etcdConfig: etcdConfig, operationLoader: operationLoader}
	g.ClusterNodeCache = ClusterNodeCache{etcdConfig: etcdConfig}
	g.UpgradeCache = UpgradeCache{etcdConfig: etcdConfig}
	g.IDNS = LocalRamDns{etcdConfig: etcdConfig}
	return nil
}

func (g *LocalRam) SyncFromDatabase() error {
	if err := g.UserRoleCache.initialization(); err != nil {
		return err
	}
	if err := g.ClusterNodeCache.initialization(); err != nil {
		return err
	}
	if err := g.OperationCache.initialization(&g.ClusterNodeCache); err != nil {
		return err
	}
	return g.UpgradeCache.initialization()
}

func (g LocalRam) SaveOrUpdateServerRuntimeConfig(nodeId string, config server.Config) error {
	NodeId = nodeId
	serverRuntimeConfigCollection[nodeId] = config
	return nil
}

func (g LocalRam) GetServerRuntimeConfig(serverId string) server.Config {
	return serverRuntimeConfigCollection[serverId]
}

func (g *LocalRam) SaveOrUpdateClientRuntimeConfig(clientId string, config client.Config) error {
	NodeId = clientId
	clientRuntimeConfigCollection[clientId] = config
	return nil
}

func (g LocalRam) GetClientRuntimeConfig(clientId string) client.Config {
	return clientRuntimeConfigCollection[clientId]
}
