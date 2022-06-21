package cache

import (
	"encoding/json"
	"fmt"
	"k8s-installer/pkg/coredns"

	etcdConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
)

var etcdDataTree = map[string]string{
	"cluster":                    "/clusters/",
	"node":                       "/nodes/",
	"operation":                  "/operations/",
	"user":                       "/users/",
	"role":                       "/roles/",
	"node-cluster-relation-ship": "/nc-ship/",
	"upgrade-version":            "/upgrade-version/",
	"upgrade-plan":               "/upgrade-plan/",
	"license":                    "/license/",
	"template":                   "/template",
	"region":                     "/regions/",
	"top-level-domain":           "/domain/top-level-domain/",
	"sub-domain":                 "/domain/sub-domain/",
}

func createOrUpdateTopLevelDomainToDB(domain coredns.TopLevelDomain, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["top-level-domain"] + domain.Domain
	return commandCreate(path, domain, config)
}

func createOrUpdateSubDomainToDB(domain coredns.DNSDomain, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["sub-domain"] + fmt.Sprintf("%s/%s", domain.TopLevelDomain, domain.Domain)
	return commandCreate(path, domain, config)
}

func createOrUpdateClusterUpgradePlanToDB(plan schema.UpgradePlan, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["upgrade-plan"] + plan.Id
	return commandCreate(path, plan, config)
}

func createOrUpdateClusterNodeRelationToDB(clusterId string, data map[string]byte, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["node-cluster-relation-ship"] + clusterId
	return commandCreate(path, data, config)
}

func createOrUpdateRoleToDB(roleId string, role schema.Role, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["role"] + roleId
	return commandCreate(path, role, config)
}

func createOrUpdateUserToDB(userId string, user schema.User, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["user"] + user.Username
	return commandCreate(path, user, config)
}

func createOrUpdateClusterToDB(clusterId string, cluster schema.Cluster, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["cluster"] + clusterId
	return commandCreate(path, cluster, config)
}

func createOrUpdateNodeInformationToDB(nodeId string, node schema.NodeInformation, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["node"] + nodeId
	return commandCreate(path, node, config)
}

func createOrUpdadeLicense(license schema.LicenseInfo, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["license"]
	return commandCreate(path, license, config)
}

func createOrUpdateTemplate(template *schema.ClusterTemplate, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["template"]
	return commandCreate(path, template, config)
}

func commandCreate(path string, objToSave interface{}, config etcdConfig.EtcdConfig) error {
	data, err := json.Marshal(&objToSave)
	if err != nil {
		return err
	}
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	if err := client.Put(path, string(data)); err != nil {
		return err
	}
	return nil
}

func createOrUpdateOperationToDB(operationId string, operation schema.Operation, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["operation"] + operationId
	return commandCreate(path, operation, config)
}

func createOrUpdateUpgradeVersionToDB(versionName string, version schema.UpgradableVersion, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["upgrade-version"] + versionName
	return commandCreate(path, version, config)
}

func getTopLevelDomainCollectionFromDB(config etcdConfig.EtcdConfig) (schema.TopLevelDomainCollection, error) {
	var results map[string]coredns.TopLevelDomain
	data, err := commonQueryMap(etcdDataTree["top-level-domain"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getAllSubDomainCollectionFromDB(config etcdConfig.EtcdConfig) (schema.SubDomainCollection, error) {
	var results map[string]coredns.DNSDomain
	data, err := commonQueryMap(etcdDataTree["sub-domain"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getSubDomainCollectionFromDB(topLevelDomain string, config etcdConfig.EtcdConfig) (schema.SubDomainCollection, error) {
	var results map[string]coredns.DNSDomain
	data, err := commonQueryMap(etcdDataTree["sub-domain"]+fmt.Sprintf("%s/", topLevelDomain), config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUpgradePlanCollectionFromDB(config etcdConfig.EtcdConfig) (schema.UpgradePlanCollection, error) {
	var results map[string]schema.UpgradePlan
	data, err := commonQueryMap(etcdDataTree["upgrade-plan"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUpgradeVersionCollectionFromDB(config etcdConfig.EtcdConfig) (schema.UpgradeVersionCollection, error) {
	var results map[string]schema.UpgradableVersion
	data, err := commonQueryMap(etcdDataTree["upgrade-version"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUpgradePlanListFromDB(config etcdConfig.EtcdConfig) ([]schema.UpgradePlan, error) {
	var results []schema.UpgradePlan
	data, err := commonQuery(etcdDataTree["upgrade-plan"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUpgradeVersionListFromDB(config etcdConfig.EtcdConfig) ([]schema.UpgradableVersion, error) {
	var results []schema.UpgradableVersion
	data, err := commonQuery(etcdDataTree["upgrade-version"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getClusterCollectionFromDB(config etcdConfig.EtcdConfig) (schema.ClusterCollection, error) {
	var results map[string]schema.Cluster
	data, err := commonQueryMap(etcdDataTree["cluster"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUpgradePlanFromDB(planId string, config etcdConfig.EtcdConfig) (*schema.UpgradePlan, error) {
	var results []schema.UpgradePlan
	data, err := commonQuery(etcdDataTree["upgrade-plan"]+planId, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getClusterFromDB(clusterId string, config etcdConfig.EtcdConfig) (*schema.Cluster, error) {
	var results []schema.Cluster
	data, err := commonQuery(etcdDataTree["cluster"]+clusterId, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getClusterWithShortIdFromDB(clusterId string, config etcdConfig.EtcdConfig) (*schema.Cluster, error) {
	var results []schema.Cluster
	data, err := commonQuery(etcdDataTree["cluster"]+clusterId, config, true)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getClusterNodeRelationCollectionFromDB(config etcdConfig.EtcdConfig) (schema.ClusterNodeRelationShipCollection, error) {
	var results map[string]map[string]byte
	data, err := commonQueryMap(etcdDataTree["node-cluster-relation-ship"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getClusterNodeRelationFromDB(clusterId string, config etcdConfig.EtcdConfig) (map[string]byte, error) {
	var results []map[string]byte
	data, err := commonQuery(etcdDataTree["node-cluster-relation-ship"]+clusterId, config, false)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	if len(results) > 0 {
		return results[0], nil
	} else {
		return map[string]byte{}, nil
	}
}

func getNodeInformationCollectionFromDB(config etcdConfig.EtcdConfig) (schema.NodeInformationCollection, error) {
	var results map[string]schema.NodeInformation
	data, err := commonQueryMap(etcdDataTree["node"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getNodeInformationCollectionFromDBPage(config etcdConfig.EtcdConfig, pageIndex, pageSize int64) (schema.NodeInformationCollection, int64, error) {
	var results map[string]schema.NodeInformation
	data, totalPage, err := commonQueryMapPage(etcdDataTree["node-cluster-relation-ship"], config, pageIndex, pageSize)
	if err != nil {
		return nil, 0, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, 0, errParse
	}
	return results, totalPage, nil
}

func getTopLevelDomainFromDB(domain string, config etcdConfig.EtcdConfig) (*coredns.TopLevelDomain, error) {
	var results []coredns.TopLevelDomain
	data, err := commonQuery(etcdDataTree["top-level-domain"]+domain, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getSubDomainFromDB(topLevelDomain, domain string, config etcdConfig.EtcdConfig) (*coredns.DNSDomain, error) {
	var results []coredns.DNSDomain
	data, err := commonQuery(etcdDataTree["sub-domain"]+fmt.Sprintf("%s/%s", topLevelDomain, domain), config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getSubDomainOfTldFromDB(topLevelDomain string, config etcdConfig.EtcdConfig) (schema.SubDomainCollection, error) {
	var results map[string]coredns.DNSDomain
	data, err := commonQueryMap(etcdDataTree["sub-domain"]+fmt.Sprintf("%s/", topLevelDomain), config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getNodeInformationFromDB(nodeId string, config etcdConfig.EtcdConfig) (*schema.NodeInformation, error) {
	var results []schema.NodeInformation
	data, err := commonQuery(etcdDataTree["node"]+nodeId, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getUpgradeVersionFromDB(versionName string, config etcdConfig.EtcdConfig) (*schema.UpgradableVersion, error) {
	var results []schema.UpgradableVersion
	data, err := commonQuery(etcdDataTree["upgrade-version"]+versionName, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getOperationFromDB(operationId string, config etcdConfig.EtcdConfig) (*schema.Operation, error) {
	var results []schema.Operation
	data, err := commonQuery(etcdDataTree["operation"]+operationId, config, false)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getOperationCollectionFromDB(config etcdConfig.EtcdConfig) (schema.OperationCollection, error) {
	var results map[string]schema.Operation
	data, err := commonQueryMap(etcdDataTree["operation"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getUserListFromDB(config etcdConfig.EtcdConfig) (schema.UserCollection, error) {
	var users map[string]schema.User
	data, err := commonQueryMap(etcdDataTree["user"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &users)
	if errParse != nil {
		return nil, errParse
	}

	roles, errRole := getRoleListFromDB(config)
	if errRole != nil {
		log.Errorf("Failed to get role list due to error %s", errRole.Error())
		return nil, err
	}

	// find out related role
	for userIndex, user := range users {
		for roleIndex, userRole := range user.Roles {
			if val, ok := roles[userRole.Id]; ok {
				users[userIndex].Roles[roleIndex] = val
			} else {
				log.Warnf("Failed to find role id %s of user %s", userRole.Id, user.Username)
			}
		}
	}

	return users, nil
}

func getRoleListFromDB(config etcdConfig.EtcdConfig) (schema.RoleCollection, error) {
	var results map[string]schema.Role
	data, err := commonQueryMap(etcdDataTree["role"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func getLicenseFromDB(config etcdConfig.EtcdConfig) (*schema.LicenseInfo, error) {
	var results []schema.LicenseInfo
	data, err := commonQuery(etcdDataTree["license"], config, true)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func getTemplateFromDB(config etcdConfig.EtcdConfig) (*schema.ClusterTemplate, error) {
	var results []schema.ClusterTemplate
	data, err := commonQuery(etcdDataTree["template"], config, true)
	if err != nil {
		return nil, err
	}

	if len(data) > 0 {
		if err = json.Unmarshal(data, &results); err != nil {
			return nil, err
		}
		if len(results) > 0 {
			return &results[0], err
		}
	}
	return nil, nil
}

func deleteTopLevelDomainFromDB(domain string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["top-level-domain"]+domain, config)
}

func deleteSubDomainFromDB(topLevelDomain, domain string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["sub-domain"]+fmt.Sprintf("%s/%s", topLevelDomain, domain), config)
}

func deleteUpgradePlanFromDB(planId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["upgrade-plan"]+planId, config)
}

func deleteUserFromDB(userId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["user"]+userId, config)
}

func deleteRoleFromDB(roleId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["role"]+roleId, config)
}

func deleteNodeInformationFromDB(nodeId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["node"]+nodeId, config)
}

func deleteClusterCollectionFromDB(clusterId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["cluster"]+clusterId, config)
}

func deleteOperationFromDB(operationId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["operation"]+operationId, config)
}

func deleteClusterNodeRelation(clusterId string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["node-cluster-relation-ship"]+clusterId, config)
}

func deleteUpgradeVersion(version string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["upgrade-version"]+version, config)
}

func commandDelete(path string, config etcdConfig.EtcdConfig) error {
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	return client.Delete(path)
}

func commonQuery(path string, config etcdConfig.EtcdConfig, withPrefix bool) ([]byte, error) {
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	queryResult, err := client.Get(path, withPrefix)
	if err != nil {
		return nil, err
	}
	return util.UnitToSlice(queryResult.Kvs), nil
}

func commonQueryPage(path string, index, pageSize int64, config etcdConfig.EtcdConfig) ([]byte, int64, error) {
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	queryResult, totalPage, err := client.GetWithPage(path, index, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return util.UnitToSlice(queryResult.Kvs), totalPage, nil
}

func commonQueryMapPage(path string, config etcdConfig.EtcdConfig, index, pageSize int64) ([]byte, int64, error) {
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	queryResult, totalPage, err := client.GetWithPage(path, index, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return util.ToMap(queryResult.Kvs), totalPage, nil
}

func commonQueryMap(path string, config etcdConfig.EtcdConfig, withPrefix bool) ([]byte, error) {
	client := EtcdClient{
		EtcdClientConfig: config,
	}
	queryResult, err := client.Get(path, withPrefix)
	if err != nil {
		return nil, err
	}
	return util.ToMap(queryResult.Kvs), nil
}

func loadAllDataFromETCD(config etcdConfig.EtcdConfig) (nodeInformationCollection schema.NodeInformationCollection,
	clusterCollection schema.ClusterCollection,
	operationCollection schema.OperationCollection,
	userCollection schema.UserCollection,
	roleCollection schema.RoleCollection,
	clusterNodeRelationshipCollection schema.ClusterNodeRelationShipCollection,
	err error) {

	nodeInformationCollection, err = getNodeInformationCollectionFromDB(config)
	if err != nil {
		return
	}
	clusterCollection, err = getClusterCollectionFromDB(config)
	if err != nil {
		return
	}
	operationCollection, err = getOperationCollectionFromDB(config)
	if err != nil {
		return
	}
	userCollection, err = getUserListFromDB(config)
	if err != nil {
		return
	}
	roleCollection, err = getRoleListFromDB(config)
	if err != nil {
		return
	}
	clusterNodeRelationshipCollection, err = getClusterNodeRelationCollectionFromDB(config)
	return
}

func createOrUpdateRegionToDB(region schema.Region, config etcdConfig.EtcdConfig) error {
	path := etcdDataTree["region"] + region.ID
	return commandCreate(path, region, config)
}

func getRegionListFromDB(config etcdConfig.EtcdConfig) (schema.RegionCollection, error) {
	var results map[string]schema.Region
	data, err := commonQueryMap(etcdDataTree["region"], config, true)
	if err != nil {
		return nil, err
	}
	errParse := json.Unmarshal(data, &results)
	if errParse != nil {
		return nil, errParse
	}
	return results, nil
}

func deleteRegionFromDB(regionID string, config etcdConfig.EtcdConfig) error {
	return commandDelete(etcdDataTree["region"]+regionID, config)
}
