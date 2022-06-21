package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/schema"
)

type RegionCache struct {
	etcdConfig       etcdClientConfig.EtcdConfig
	regionCollection schema.RegionCollection
}

func (r *RegionCache) CreateOrUpdateRegion(region schema.Region) error {
	//TODO: add impl
	return nil
}

func (r *RegionCache) GetRegionList() (schema.RegionCollection, error) {
	//TODO: add impl
	return schema.RegionCollection{}, nil
}

func (r *RegionCache) GetRegion(regionID string) (*schema.Region, error) {
	//TODO: add impl
	return nil, nil
}

func (r *RegionCache) DeleteRegion(regionID string) error {
	//TODO: add impl
	return nil
}
