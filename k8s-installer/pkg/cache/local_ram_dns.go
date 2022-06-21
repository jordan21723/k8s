package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/coredns"
	"k8s-installer/schema"
)

type LocalRamDns struct {
	clusterCollection schema.ClusterCollection
	etcdConfig        etcdClientConfig.EtcdConfig
}

func (l LocalRamDns) CreateOrUpdateTopLevelDomain(domain coredns.TopLevelDomain) error {
	panic("implement me")
}

func (l LocalRamDns) GetTopLevelDomainList() (schema.TopLevelDomainCollection, error) {
	panic("implement me")
}

func (l LocalRamDns) GetTopLevelDomain(domain string) (*coredns.TopLevelDomain, error) {
	panic("implement me")
}

func (l LocalRamDns) DeleteTopLevelDomain(domain string) error {
	panic("implement me")
}

func (l LocalRamDns) CreateOrUpdateSubDomain(domain coredns.DNSDomain) error {
	panic("implement me")
}

func (l LocalRamDns) GetSubDomainList(topLevelDomain string) (schema.SubDomainCollection, error) {
	panic("implement me")
}

func (l LocalRamDns) GetAllSubDomainList() (schema.SubDomainCollection, error) {
	panic("implement me")
}

func (l LocalRamDns) GetSubDomain(topLevelDomain, domain string) (*coredns.DNSDomain, error) {
	panic("implement me")
}

func (l LocalRamDns) DeleteSubDomain(topLevelDomain, domain string) error {
	panic("implement me")
}

func (l LocalRamDns) GetSubDomainOfTls(topLevelDomain string) (schema.SubDomainCollection, error) {
	panic("implement me")
}
