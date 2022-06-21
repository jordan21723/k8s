package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type OperationNoCache struct {
	etcdConfig              etcdClientConfig.EtcdConfig
	inputClusterNodeNoCache ClusterNodeNoCache
	operationLoader         func(operation schema.Operation, cluster schema.Cluster, config server.Config, nodeCollectionList schema.NodeInformationCollection) (schema.Operation, error)
}

func (noCachedOperation *OperationNoCache) GetOperationCollection() (schema.OperationCollection, error) {
	result, err := getOperationCollectionFromDB(noCachedOperation.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return result, nil
}

func (noCachedOperation *OperationNoCache) GetOperationDoNotLoadStep(operationId string) (*schema.Operation, error) {
	operation, err := getOperationFromDB(operationId, noCachedOperation.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return operation, nil
}

func (noCachedOperation *OperationNoCache) GetOperation(operationId string) (*schema.Operation, error) {
	operation, err := getOperationFromDB(operationId, noCachedOperation.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if operation == nil {
		return nil, nil
	}

	cluster, errGetCluster := noCachedOperation.inputClusterNodeNoCache.GetCluster(operation.ClusterId)
	if errGetCluster != nil {
		return nil, errGetCluster
	}

	if cluster == nil {
		return nil, nil
	}

	nodeCollection, errGetNodes := noCachedOperation.inputClusterNodeNoCache.GetNodeInformationCollection()
	if errGetNodes != nil {
		return nil, errGetNodes
	}

	if operation, err := noCachedOperation.operationLoader(*operation, *cluster,
		serverRuntimeConfigCollection[NodeId],
		nodeCollection); err != nil {
		log.Errorf("Failed to initial operation collection due to error: %s", err)
		return nil, err
	} else {
		return &operation, nil
	}
}

func (noCachedOperation *OperationNoCache) SaveOrUpdateOperationCollection(operationId string, operation schema.Operation) error {
	return createOrUpdateOperationToDB(operationId, operation, noCachedOperation.etcdConfig)
}

func (noCachedOperation *OperationNoCache) SetOperationStatus(operationId, status string) error {
	operation, err := getOperationFromDB(operationId, noCachedOperation.etcdConfig)
	if err != nil {
		return err
	}
	operation.Status = status
	return createOrUpdateOperationToDB(operationId, *operation, noCachedOperation.etcdConfig)
}

func (noCachedOperation *OperationNoCache) DeleteOperation(operationId string) error {
	return deleteOperationFromDB(operationId, noCachedOperation.etcdConfig)
}

func (noCachedOperation *OperationNoCache) GetClusterOperationDoNotLoadStep(cluster schema.Cluster) ([]schema.Operation, error) {
	if len(cluster.ClusterOperationIDs) == 0 {
		return []schema.Operation{}, nil
	}
	var results []schema.Operation
	for i := len(cluster.ClusterOperationIDs) - 1; i >= 0; i-- {
		operation, err := noCachedOperation.GetOperationDoNotLoadStep(cluster.ClusterOperationIDs[i])
		if err != nil {
			return nil, err
		}
		if operation == nil {
			log.Warn("Operation with id %s not found in db", cluster.ClusterOperationIDs[i])
			continue
		}
		results = append(results, *operation)
	}
	return results, nil
}

func (noCachedOperation *OperationNoCache) GetClusterOperation(cluster schema.Cluster) ([]schema.Operation, error) {
	if len(cluster.ClusterOperationIDs) == 0 {
		return []schema.Operation{}, nil
	}
	var results []schema.Operation
	for i := len(cluster.ClusterOperationIDs) - 1; i >= 0; i-- {
		operation, err := noCachedOperation.GetOperation(cluster.ClusterOperationIDs[i])
		if err != nil {
			return nil, err
		}
		if operation == nil {
			log.Warn("Operation with id %s not found in db", cluster.ClusterOperationIDs[i])
			continue
		}
		results = append(results, *operation)
	}
	return results, nil
}
