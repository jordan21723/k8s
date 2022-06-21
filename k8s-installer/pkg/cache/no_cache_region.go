package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/schema"
)

type RegionNoCache struct {
	etcdConfig etcdClientConfig.EtcdConfig
}

func (r *RegionNoCache) initialization() error {
	return nil
}

func (r *RegionNoCache) CreateOrUpdateRegion(region schema.Region) error {
	return createOrUpdateRegionToDB(region, r.etcdConfig)
}

func (r *RegionNoCache) GetRegionList() (schema.RegionCollection, error) {
	return getRegionListFromDB(r.etcdConfig)
}

func (r *RegionNoCache) GetRegion(regionID string) (*schema.Region, error) {
	regionList, err := getRegionListFromDB(r.etcdConfig)
	if err != nil {
		return nil, err
	}
	for _, region := range regionList {
		if region.ID == regionID {
			return &schema.Region{
				ID:                region.ID,
				Name:              region.Name,
				CreationTimestamp: region.CreationTimestamp,
				UpdateTimestamp:   region.UpdateTimestamp.DeepCopy(),
			}, nil
		}
	}
	return nil, nil
}

func (r *RegionNoCache) DeleteRegion(regionID string) error {
	return deleteRegionFromDB(regionID, r.etcdConfig)
}
