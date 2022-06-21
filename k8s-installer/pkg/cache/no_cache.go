package cache

import (
	"k8s-installer/pkg/config/client"
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/schema"
)

/*
no-cache simply get every thing
*/

type NoCache struct {
	UserRoleNoCache
	ClusterNodeNoCache
	OperationNoCache
	UpgradeVersionNoCache
	RegionNoCache
	IDNS
	//etcdConfig                        etcdClientConfig.EtcdConfig
}

func (n *NoCache) initialization(config etcdClientConfig.EtcdConfig, operationLoader func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)) error {
	n.UserRoleNoCache = UserRoleNoCache{etcdConfig: config}
	n.ClusterNodeNoCache = ClusterNodeNoCache{etcdConfig: config}
	n.OperationNoCache = OperationNoCache{etcdConfig: config, inputClusterNodeNoCache: n.ClusterNodeNoCache, operationLoader: operationLoader}
	n.UpgradeVersionNoCache = UpgradeVersionNoCache{etcdConfig: config}
	n.RegionNoCache = RegionNoCache{etcdConfig: config}
	n.IDNS = NoCacheDns{etcdConfig: config}
	return nil
}

func (n NoCache) SaveOrUpdateServerRuntimeConfig(nodeId string, config server.Config) error {
	NodeId = config.ServerId
	serverRuntimeConfigCollection[nodeId] = config
	return nil
}

func (n NoCache) GetServerRuntimeConfig(serverId string) server.Config {
	return serverRuntimeConfigCollection[serverId]
}

func (g *NoCache) SaveOrUpdateClientRuntimeConfig(clientId string, config client.Config) error {
	NodeId = config.ClientId
	clientRuntimeConfigCollection[clientId] = config
	return nil
}

func (g NoCache) GetClientRuntimeConfig(clientId string) client.Config {
	return clientRuntimeConfigCollection[clientId]
}

func (n NoCache) SyncFromDatabase() error {
	return nil
}
