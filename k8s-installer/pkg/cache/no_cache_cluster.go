package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type ClusterNodeNoCache struct {
	etcdConfig etcdClientConfig.EtcdConfig
}

func (noCachedClusterNode *ClusterNodeNoCache) GetAvailableNodesMap(pageIndex, pageSize int64,
	filter func(information schema.NodeInformation) *schema.NodeInformation) (schema.NodeInformationCollection, int64, error) {
	fullNodeList, err := getNodeInformationCollectionFromDB(noCachedClusterNode.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}
	return getAvailableNodesMap(fullNodeList, pageIndex, pageSize, filter)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetNodeInformationCollection() (schema.NodeInformationCollection, error) {
	result, err := getNodeInformationCollectionFromDB(noCachedClusterNode.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return result, nil
}

func (noCachedClusterNode *ClusterNodeNoCache) SaveOrUpdateNodeInformation(nodeId string, information schema.NodeInformation) error {
	return createOrUpdateNodeInformationToDB(nodeId, information, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) DeleteNodeInformation(nodeId string) error {
	return deleteNodeInformationFromDB(nodeId, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetNodeInformation(nodeId string) (*schema.NodeInformation, error) {
	result, err := getNodeInformationFromDB(nodeId, noCachedClusterNode.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return result, nil
}

func (noCachedClusterNode *ClusterNodeNoCache) SetNodeStatus(nodeId, status string) error {
	information, err := getNodeInformationFromDB(nodeId, noCachedClusterNode.etcdConfig)
	if err != nil {
		return err
	}
	information.Status = status
	return createOrUpdateNodeInformationToDB(nodeId, *information, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetClusterCollection() (schema.ClusterCollection, error) {
	return getClusterCollectionFromDB(noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) SaveOrUpdateClusterCollection(clusterId string, cluster schema.Cluster) error {
	return createOrUpdateClusterToDB(clusterId, cluster, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetCluster(clusterId string) (*schema.Cluster, error) {
	result, err := getClusterFromDB(clusterId, noCachedClusterNode.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return result, nil
}

func (noCachedClusterNode *ClusterNodeNoCache) GetClusterWithShortId(shortClusterId string) (*schema.Cluster, error) {
	result, err := getClusterWithShortIdFromDB(shortClusterId, noCachedClusterNode.etcdConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return result, nil
}

func (noCachedClusterNode *ClusterNodeNoCache) DeleteClusterCollection(clusterId string) error {
	return deleteClusterCollectionFromDB(clusterId, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) AddClusterNodeRelationShip(clusterId string, nodes []string) error {
	if len(nodes) <= 0 {
		return nil
	}

	if dataToSave, err := noCachedClusterNode.GetClusterNodeRelationShip(clusterId); err != nil {
		return err
	} else {
		for _, node := range nodes {
			dataToSave[node] = 0
		}
		if err := createOrUpdateClusterNodeRelationToDB(clusterId, dataToSave, noCachedClusterNode.etcdConfig); err != nil {
			return err
		}
	}

	return nil
}

func (noCachedClusterNode *ClusterNodeNoCache) GetClusterNodeRelationShip(clusterId string) (map[string]byte, error) {
	return getClusterNodeRelationFromDB(clusterId, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) RemoveClusterNodeRelationShip(clusterId string, nodes []string) error {
	if len(nodes) <= 0 {
		return nil
	}

	if remainList, err := getClusterNodeRelationFromDB(clusterId, noCachedClusterNode.etcdConfig); err != nil {
		return err
	} else {
		for _, node := range nodes {
			delete(remainList, node)
		}
		// save the remain list to db
		if err := createOrUpdateClusterNodeRelationToDB(clusterId, remainList, noCachedClusterNode.etcdConfig); err != nil {
			return err
		}
	}
	return nil
}

func (noCachedClusterNode *ClusterNodeNoCache) RemoveAllClusterNodeRelationShip(clusterId string) error {
	if err := deleteClusterNodeRelation(clusterId, noCachedClusterNode.etcdConfig); err != nil {
		return err
	}
	return nil
}

func (noCachedClusterNode *ClusterNodeNoCache) GetNodeInformationWithCluster(clusterId string) ([]schema.NodeInformation, error) {
	var results []schema.NodeInformation
	relationship, err := getClusterNodeRelationFromDB(clusterId, noCachedClusterNode.etcdConfig)
	if err != nil {
		return nil, err
	}
	if len(relationship) == 0 {
		return nil, nil
	}
	for nodeId := range relationship {
		if nodeInfo, err := getNodeInformationFromDB(nodeId, noCachedClusterNode.etcdConfig); err != nil {
			return nil, err
		} else {
			if nodeInfo != nil {
				results = append(results, *nodeInfo)
			}
		}
	}
	return results, nil
}

func (noCachedClusterNode *ClusterNodeNoCache) GetNodeInformationCollectionWithPage(pageIndex, pageSize int64) (schema.NodeInformationCollection, int64, error) {
	return getNodeInformationCollectionFromDBPage(noCachedClusterNode.etcdConfig, pageIndex, pageSize)
}

func (noCachedClusterNode *ClusterNodeNoCache) CreateOrUpdadeLicense(license schema.LicenseInfo) error {
	return createOrUpdadeLicense(license, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetLicense() (*schema.LicenseInfo, error) {
	return getLicenseFromDB(noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) SaveClusterTemplate(template *schema.ClusterTemplate) error {
	return createOrUpdateTemplate(template, noCachedClusterNode.etcdConfig)
}

func (noCachedClusterNode *ClusterNodeNoCache) GetTemplate() (*schema.ClusterTemplate, error) {
	return getTemplateFromDB(noCachedClusterNode.etcdConfig)
}
