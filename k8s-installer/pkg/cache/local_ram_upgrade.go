package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/schema"
	"sort"
)

type UpgradeCache struct {
	etcdConfig               etcdClientConfig.EtcdConfig
	upgradeVersionCollection schema.UpgradeVersionCollection
	upgradePlanCollection    schema.UpgradePlanCollection
}

func (cachedUpgradeVersion *UpgradeCache) initialization() error {
	var err error
	cachedUpgradeVersion.upgradeVersionCollection, err = getUpgradeVersionCollectionFromDB(cachedUpgradeVersion.etcdConfig)
	if err != nil {
		return err
	}
	cachedUpgradeVersion.upgradePlanCollection, err = getUpgradePlanCollectionFromDB(cachedUpgradeVersion.etcdConfig)
	if err != nil {
		return err
	}
	return nil
}

func (cachedUpgradeVersion *UpgradeCache) GetUpgradeVersionList() ([]schema.UpgradableVersion, error) {
	var versionKeys []string
	var results []schema.UpgradableVersion
	for key := range cachedUpgradeVersion.upgradeVersionCollection {
		versionKeys = append(versionKeys, key)
	}
	sort.Strings(versionKeys)
	for _, key := range versionKeys {
		results = append(results, cachedUpgradeVersion.upgradeVersionCollection[key])
	}
	return results, nil
}

func (cachedUpgradeVersion *UpgradeCache) GetUpgradeVersionCollection() (schema.UpgradeVersionCollection, error) {
	return cachedUpgradeVersion.upgradeVersionCollection, nil
}

func (cachedUpgradeVersion *UpgradeCache) CreateOrUpdateUpgradeVersion(version schema.UpgradableVersion) error {
	if err := createOrUpdateUpgradeVersionToDB(version.Name, version, cachedUpgradeVersion.etcdConfig); err != nil {
		return err
	}
	cachedUpgradeVersion.upgradeVersionCollection[version.Name] = version
	return nil
}

func (cachedUpgradeVersion *UpgradeCache) DeleteUpgradeVersion(version string) error {
	if err := deleteUpgradeVersion(version, cachedUpgradeVersion.etcdConfig); err != nil {
		return err
	}
	delete(cachedUpgradeVersion.upgradeVersionCollection, version)
	return nil
}
func (cachedUpgradeVersion *UpgradeCache) GetUpgradeVersion(versionName string) (*schema.UpgradableVersion, error) {
	if found, exists := cachedUpgradeVersion.upgradeVersionCollection[versionName]; exists {
		return &found, nil
	} else {
		return nil, nil
	}
}

func (cachedUpgradeVersion *UpgradeCache) CreateOrUpdateUpgradePlan(plan schema.UpgradePlan) error {
	if err := createOrUpdateClusterUpgradePlanToDB(plan, cachedUpgradeVersion.etcdConfig); err != nil {
		return err
	}
	cachedUpgradeVersion.upgradePlanCollection[plan.Id] = plan
	return nil
}

func (cachedUpgradeVersion *UpgradeCache) GetUpgradePlanCollection() (schema.UpgradePlanCollection, error) {
	return cachedUpgradeVersion.upgradePlanCollection, nil
}

func (cachedUpgradeVersion *UpgradeCache) QueryUpgradePlan(filter func(information schema.UpgradePlan) *schema.UpgradePlan) ([]schema.UpgradePlan, error) {
	var results []schema.UpgradePlan
	for _, plan := range cachedUpgradeVersion.upgradePlanCollection {
		if filter != nil {
			passedFilterPlan := filter(plan)
			if passedFilterPlan != nil {
				results = append(results, *passedFilterPlan)
			}
		} else {
			results = append(results, plan)
		}
	}
	return results, nil
}

func (cachedUpgradeVersion *UpgradeCache) GetUpgradePlanList() ([]schema.UpgradePlan, error) {
	var results []schema.UpgradePlan
	for _, plan := range cachedUpgradeVersion.upgradePlanCollection {
		results = append(results, plan)
	}
	return results, nil
}

func (cachedUpgradeVersion *UpgradeCache) GetUpgradePlan(id string) (*schema.UpgradePlan, error) {
	if found, exists := cachedUpgradeVersion.upgradePlanCollection[id]; exists {
		return &found, nil
	} else {
		return nil, nil
	}
}

func (cachedUpgradeVersion *UpgradeCache) DeleteUpgradePlan(id string) error {
	if err := deleteUpgradePlanFromDB(id, cachedUpgradeVersion.etcdConfig); err != nil {
		return nil
	}
	delete(cachedUpgradeVersion.upgradePlanCollection, id)
	return nil
}
