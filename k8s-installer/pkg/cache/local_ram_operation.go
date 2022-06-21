package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type OperationCache struct {
	etcdConfig          etcdClientConfig.EtcdConfig
	operationCollection schema.OperationCollection
	operationLoader     func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)
}

func (cachedOperation *OperationCache) initialization(cachedClusterNode *ClusterNodeCache) error {
	var err error
	cachedOperation.operationCollection, err = getOperationCollectionFromDB(cachedOperation.etcdConfig)
	if err != nil {
		return err
	}
	for key, operation := range cachedOperation.operationCollection {
		if _, exists := cachedClusterNode.clusterCollection[operation.ClusterId]; exists {
			if operation, err := cachedOperation.operationLoader(operation, cachedClusterNode.clusterCollection[operation.ClusterId],
				serverRuntimeConfigCollection[NodeId],
				cachedClusterNode.nodeInformationCollection); err != nil {
				log.Fatalf("Failed to initial operation collection due to error: %s", err)
			} else {
				cachedOperation.operationCollection[key] = operation
			}
		}
	}
	return nil
}

func (cachedOperation *OperationCache) GetOperationCollection() (schema.OperationCollection, error) {
	return cachedOperation.operationCollection, nil
}

func (cachedOperation *OperationCache) SaveOrUpdateOperationCollection(operationId string, operation schema.Operation) error {
	if err := createOrUpdateOperationToDB(operationId, operation, cachedOperation.etcdConfig); err != nil {
		return err
	}
	cachedOperation.operationCollection[operationId] = operation
	return nil
}

func (cachedOperation *OperationCache) SetOperationStatus(operationId, status string) error {
	panic("implement me")
}

func (cachedOperation *OperationCache) DeleteOperation(operationId string) error {
	panic("implement me")
}

func (cachedOperation *OperationCache) GetOperation(operationId string) (*schema.Operation, error) {
	if _, exists := cachedOperation.operationCollection[operationId]; exists {
		operation := cachedOperation.operationCollection[operationId]
		return &operation, nil
	}
	return nil, nil
}

func (cachedOperation *OperationCache) GetClusterOperationDoNotLoadStep(cluster schema.Cluster) ([]schema.Operation, error) {
	return nil, nil
}

func (cachedOperation *OperationCache) GetClusterOperation(cluster schema.Cluster) ([]schema.Operation, error) {
	if len(cluster.ClusterOperationIDs) == 0 {
		return []schema.Operation{}, nil
	}
	var results []schema.Operation
	for i := len(cluster.ClusterOperationIDs) - 1; i >= 0; i-- {
		operation, err := cachedOperation.GetOperation(cluster.ClusterOperationIDs[i])
		if err != nil {
			return nil, err
		}
		if operation == nil {
			return []schema.Operation{}, nil
		}
		results = append(results, *operation)
	}
	return results, nil
}
