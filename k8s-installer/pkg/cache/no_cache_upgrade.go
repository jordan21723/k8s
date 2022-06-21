package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/schema"
)

type UpgradeVersionNoCache struct {
	etcdConfig etcdClientConfig.EtcdConfig
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradeVersionList() ([]schema.UpgradableVersion, error) {
	return getUpgradeVersionListFromDB(noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradeVersionCollection() (schema.UpgradeVersionCollection, error) {
	return getUpgradeVersionCollectionFromDB(noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) CreateOrUpdateUpgradeVersion(upgradeVersion schema.UpgradableVersion) error {
	return createOrUpdateUpgradeVersionToDB(upgradeVersion.Name, upgradeVersion, noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) DeleteUpgradeVersion(version string) error {
	return deleteUpgradeVersion(version, noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradeVersion(versionName string) (*schema.UpgradableVersion, error) {
	if found, err := getUpgradeVersionFromDB(versionName, noCachedUpgradeVersion.etcdConfig); err != nil {
		return nil, err
	} else {
		return found, nil
	}
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) CreateOrUpdateUpgradePlan(plan schema.UpgradePlan) error {
	if err := createOrUpdateClusterUpgradePlanToDB(plan, noCachedUpgradeVersion.etcdConfig); err != nil {
		return err
	}
	return nil
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradePlanCollection() (schema.UpgradePlanCollection, error) {
	return getUpgradePlanCollectionFromDB(noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) QueryUpgradePlan(filter func(information schema.UpgradePlan) *schema.UpgradePlan) ([]schema.UpgradePlan, error) {
	var results []schema.UpgradePlan
	if plans, err := noCachedUpgradeVersion.GetUpgradePlanCollection(); err != nil {
		return nil, err
	} else {
		for _, plan := range plans {
			if filter != nil {
				passedFilterPlan := filter(plan)
				if passedFilterPlan != nil {
					results = append(results, *passedFilterPlan)
				}
			} else {
				results = append(results, plan)
			}
		}
	}
	return results, nil
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradePlanList() ([]schema.UpgradePlan, error) {
	return getUpgradePlanListFromDB(noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) GetUpgradePlan(id string) (*schema.UpgradePlan, error) {
	return getUpgradePlanFromDB(id, noCachedUpgradeVersion.etcdConfig)
}

func (noCachedUpgradeVersion *UpgradeVersionNoCache) DeleteUpgradePlan(id string) error {
	return deleteUpgradePlanFromDB(id, noCachedUpgradeVersion.etcdConfig)
}
