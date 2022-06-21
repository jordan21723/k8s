package cache

import (
	"sort"

	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/schema"
)

type ClusterNodeCache struct {
	clusterCollection         schema.ClusterCollection
	nodeInformationCollection schema.NodeInformationCollection
	clusterNodeRelationShip   schema.ClusterNodeRelationShipCollection
	etcdConfig                etcdClientConfig.EtcdConfig
}

func (cachedClusterNode *ClusterNodeCache) initialization() error {
	var err error
	cachedClusterNode.clusterCollection, err = getClusterCollectionFromDB(cachedClusterNode.etcdConfig)
	if err != nil {
		return err
	}
	cachedClusterNode.nodeInformationCollection, err = getNodeInformationCollectionFromDB(cachedClusterNode.etcdConfig)
	if err != nil {
		return err
	}
	cachedClusterNode.clusterNodeRelationShip, err = getClusterNodeRelationCollectionFromDB(cachedClusterNode.etcdConfig)
	if err != nil {
		return err
	}
	return nil
}

func (cachedClusterNode *ClusterNodeCache) GetAvailableNodesMap(pageIndex, pageSize int64,
	filter func(information schema.NodeInformation) *schema.NodeInformation) (schema.NodeInformationCollection, int64, error) {
	return getAvailableNodesMap(cachedClusterNode.nodeInformationCollection, pageIndex, pageSize, filter)
}

func (cachedClusterNode *ClusterNodeCache) GetNodeInformationCollection() (schema.NodeInformationCollection, error) {
	return cachedClusterNode.nodeInformationCollection, nil
}

func (cachedClusterNode *ClusterNodeCache) SaveOrUpdateNodeInformation(nodeId string, nodeInformation schema.NodeInformation) error {
	if err := createOrUpdateNodeInformationToDB(nodeId, nodeInformation, cachedClusterNode.etcdConfig); err != nil {
		return err
	}
	cachedClusterNode.nodeInformationCollection[nodeId] = nodeInformation
	return nil
}

func (cachedClusterNode *ClusterNodeCache) DeleteNodeInformation(nodeId string) error {
	if err := deleteNodeInformationFromDB(nodeId, cachedClusterNode.etcdConfig); err != nil {
		return err
	}
	delete(cachedClusterNode.nodeInformationCollection, nodeId)
	return nil
}

func (cachedClusterNode *ClusterNodeCache) SetNodeStatus(nodeId, status string) error {
	panic("implement me")
}

func (cachedClusterNode *ClusterNodeCache) GetClusterCollection() (schema.ClusterCollection, error) {
	result, err := getClusterCollectionFromDB(cachedClusterNode.etcdConfig)
	if err != nil {
		return nil, nil
	}
	return result, nil
	//return l.clusterCollection,nil
}

func (cachedClusterNode *ClusterNodeCache) SaveOrUpdateClusterCollection(clusterId string, cluster schema.Cluster) error {
	if err := createOrUpdateClusterToDB(clusterId, cluster, cachedClusterNode.etcdConfig); err != nil {
		return err
	}
	cachedClusterNode.clusterCollection[clusterId] = cluster
	return nil
}

func (cachedClusterNode *ClusterNodeCache) DeleteClusterCollection(clusterId string) error {
	panic("implement me")
}

func (cachedClusterNode *ClusterNodeCache) AddClusterNodeRelationShip(clusterId string, nodes []string) error {

	if len(nodes) <= 0 {
		return nil
	}

	var dataToSave map[string]byte

	if _, found := cachedClusterNode.clusterNodeRelationShip[clusterId]; found {
		dataToSave = cachedClusterNode.clusterNodeRelationShip[clusterId]
	} else {
		dataToSave = map[string]byte{}
	}

	for _, node := range nodes {
		dataToSave[node] = 0
	}

	if err := createOrUpdateClusterNodeRelationToDB(clusterId, dataToSave, cachedClusterNode.etcdConfig); err != nil {
		return err
	}

	cachedClusterNode.clusterNodeRelationShip[clusterId] = dataToSave
	return nil
}

func (cachedClusterNode *ClusterNodeCache) GetClusterNodeRelationShip(clusterId string) (map[string]byte, error) {
	return cachedClusterNode.clusterNodeRelationShip[clusterId], nil
}

func (cachedClusterNode *ClusterNodeCache) RemoveClusterNodeRelationShip(clusterId string, nodes []string) error {
	remainNodes := cachedClusterNode.clusterNodeRelationShip[clusterId]
	for _, node := range nodes {
		delete(remainNodes, node)
	}
	if err := createOrUpdateClusterNodeRelationToDB(clusterId, remainNodes, cachedClusterNode.etcdConfig); err != nil {
		return err
	}
	cachedClusterNode.clusterNodeRelationShip[clusterId] = remainNodes
	return nil
}

func (cachedClusterNode *ClusterNodeCache) RemoveAllClusterNodeRelationShip(clusterId string) error {
	// save the remain list to db
	if err := deleteClusterNodeRelation(clusterId, cachedClusterNode.etcdConfig); err != nil {
		return err
	}
	delete(cachedClusterNode.nodeInformationCollection, clusterId)
	return nil
}

func (cachedClusterNode *ClusterNodeCache) GetNodeInformationWithCluster(clusterId string) ([]schema.NodeInformation, error) {
	var results []schema.NodeInformation
	if relation, found := cachedClusterNode.clusterNodeRelationShip[clusterId]; !found {
		return results, nil
	} else {
		for node := range relation {
			if nodeInfo, nodeFound := cachedClusterNode.nodeInformationCollection[node]; nodeFound {
				results = append(results, nodeInfo)
			}
		}
	}
	return results, nil
}

func (cachedClusterNode *ClusterNodeCache) GetNodeInformationCollectionWithPage(pageIndex, pageSize int64) (schema.NodeInformationCollection, int64, error) {
	nodeCount := int64(len(cachedClusterNode.nodeInformationCollection))
	if nodeCount == 0 {
		return nil, 0, nil
	}
	var nodeIdList []string
	result := schema.NodeInformationCollection{}
	for key := range cachedClusterNode.nodeInformationCollection {
		nodeIdList = append(nodeIdList, key)
	}
	sort.Strings(nodeIdList)

	totalPage, pageStart, pageEnd := pageUp(nodeCount, pageSize, pageIndex)

	if pageStart >= nodeCount {
		// page index do not exists set data to empty
		return nil, 0, nil
	} else {
		// return ranged data

		for i := pageStart; i < pageEnd; i++ {
			result[nodeIdList[i]] = cachedClusterNode.nodeInformationCollection[nodeIdList[i]]
		}
	}

	return result, totalPage, nil
}

func (cachedClusterNode *ClusterNodeCache) GetCluster(clusterId string) (*schema.Cluster, error) {
	if _, exists := cachedClusterNode.clusterCollection[clusterId]; exists {
		cluster := cachedClusterNode.clusterCollection[clusterId]
		return &cluster, nil
	}
	return nil, nil
}

func (cachedClusterNode *ClusterNodeCache) GetClusterWithShortId(shortClusterId string) (*schema.Cluster, error) {
	return nil, nil
}

func (cachedClusterNode *ClusterNodeCache) GetNodeInformation(nodeId string) (*schema.NodeInformation, error) {
	if _, exists := cachedClusterNode.nodeInformationCollection[nodeId]; exists {
		nodeInfo := cachedClusterNode.nodeInformationCollection[nodeId]
		return &nodeInfo, nil
	}
	return nil, nil
}

func (cachedClusterNode *ClusterNodeCache) CreateOrUpdadeLicense(license schema.LicenseInfo) error {
	return createOrUpdadeLicense(license, cachedClusterNode.etcdConfig)
}

func (cachedClusterNode *ClusterNodeCache) GetLicense() (*schema.LicenseInfo, error) {
	return getLicenseFromDB(cachedClusterNode.etcdConfig)
}

func (cachedClusterNode *ClusterNodeCache) SaveClusterTemplate(template *schema.ClusterTemplate) error {
	return createOrUpdateTemplate(template, cachedClusterNode.etcdConfig)
}

func (cachedClusterNode *ClusterNodeCache) GetTemplate() (*schema.ClusterTemplate, error) {
	return getTemplateFromDB(cachedClusterNode.etcdConfig)
}
