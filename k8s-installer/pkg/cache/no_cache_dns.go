package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/coredns"
	"k8s-installer/schema"
)

type NoCacheDns struct {
	clusterCollection schema.ClusterCollection
	etcdConfig        etcdClientConfig.EtcdConfig
}

func (dns NoCacheDns) CreateOrUpdateSubDomain(domain coredns.DNSDomain) error {
	return createOrUpdateSubDomainToDB(domain, dns.etcdConfig)
}

func (dns NoCacheDns) GetSubDomainList(topLevelDomain string) (schema.SubDomainCollection, error) {
	return getSubDomainCollectionFromDB(topLevelDomain, dns.etcdConfig)
}

func (dns NoCacheDns) GetAllSubDomainList() (schema.SubDomainCollection, error) {
	return getAllSubDomainCollectionFromDB(dns.etcdConfig)
}

func (dns NoCacheDns) GetSubDomain(topLevelDomain, domain string) (*coredns.DNSDomain, error) {
	return getSubDomainFromDB(topLevelDomain, domain, dns.etcdConfig)
}

func (dns NoCacheDns) GetSubDomainOfTls(topLevelDomain string) (schema.SubDomainCollection, error) {
	return getSubDomainOfTldFromDB(topLevelDomain, dns.etcdConfig)
}

func (dns NoCacheDns) DeleteSubDomain(topLevelDomain, domain string) error {
	return deleteSubDomainFromDB(topLevelDomain, domain, dns.etcdConfig)
}

func (dns NoCacheDns) CreateOrUpdateTopLevelDomain(domain coredns.TopLevelDomain) error {
	return createOrUpdateTopLevelDomainToDB(domain, dns.etcdConfig)
}

func (dns NoCacheDns) GetTopLevelDomainList() (schema.TopLevelDomainCollection, error) {
	return getTopLevelDomainCollectionFromDB(dns.etcdConfig)
}

func (dns NoCacheDns) GetTopLevelDomain(domain string) (*coredns.TopLevelDomain, error) {
	return getTopLevelDomainFromDB(domain, dns.etcdConfig)
}

func (dns NoCacheDns) DeleteTopLevelDomain(domain string) error {
	return deleteTopLevelDomainFromDB(domain, dns.etcdConfig)
}
