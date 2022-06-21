package cache

import (
	"context"
	"crypto/tls"
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"strings"
	"time"

	etcdv3 "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/pkg/transport"
)

type EtcdClient struct {
	v3Client         *etcdv3.Client
	Ctx              context.Context
	Kv               etcdv3.KV
	EtcdClientConfig etcdClientConfig.EtcdConfig
}

func (etcdClient *EtcdClient) ConnectDB() error {
	var err error
	// total connection timeout 10 secs
	etcdClient.Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	clientConfig, err := etcdClient.getClientConfig()
	if err != nil {
		return err
	}
	etcdClient.v3Client, err = etcdv3.New(*clientConfig)
	etcdClient.Kv = etcdv3.NewKV(etcdClient.v3Client)
	return nil
}

func (etcdClient *EtcdClient) CloseConnection() error {
	return etcdClient.v3Client.Close()
}

func (etcdClient *EtcdClient) getClientConfig() (*etcdv3.Config, error) {
	/*	currentCache := GetCurrentCache()
		nodeConfig := currentCache.GetServerRuntimeConfig(NodeId)*/
	switch etcdClient.EtcdClientConfig.AuthMode {
	case "tls":
		tlsConfig, err := getTlsConfig(etcdClient.EtcdClientConfig)
		if err != nil {
			return nil, err
		}
		return &etcdv3.Config{
			DialTimeout: 2 * time.Second,
			Endpoints:   strings.Split(etcdClient.EtcdClientConfig.EndPoints, " | "),
			TLS:         tlsConfig,
		}, nil
	default:
		return &etcdv3.Config{
			DialTimeout: 2 * time.Second,
			Endpoints:   strings.Split(etcdClient.EtcdClientConfig.EndPoints, " | "),
		}, nil
	}
}

func getTlsConfig(config etcdClientConfig.EtcdConfig) (*tls.Config, error) {
	tlsInfo := transport.TLSInfo{
		CertFile:      config.CertFile,
		KeyFile:       config.KeyFile,
		TrustedCAFile: config.CaCertFile,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func (etcdClient *EtcdClient) Put(etcdPath string, objToSave string) error {
	defer etcdClient.CloseConnection()
	if err := etcdClient.ConnectDB(); err != nil {
		return nil
	}
	pr, errPut := etcdClient.Kv.Put(etcdClient.Ctx, etcdPath, objToSave)
	if errPut != nil {
		return errPut
	}
	log.Debugf("Successfully put data to etcd etcd revision id: %d", pr.Header.Revision)
	return nil
	//return etcdClient.CloseConnection()
}

func (etcdClient *EtcdClient) Get(path string, withPrefix bool) (*etcdv3.GetResponse, error) {
	defer etcdClient.CloseConnection()

	if err := etcdClient.ConnectDB(); err != nil {
		return nil, err
	}

	opts := []etcdv3.OpOption{
		//etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByKey, etcdv3.SortAscend),
	}
	if withPrefix {
		opts = append(opts, etcdv3.WithPrefix())
	}

	kvs, err := etcdClient.Kv.Get(etcdClient.Ctx, path, opts...)
	if err != nil {
		return nil, err
	}
	return kvs, nil
}

func (etcdClient *EtcdClient) GetWithPage(path string, index, pageSize int64) (*etcdv3.GetResponse, int64, error) {
	defer etcdClient.CloseConnection()
	if err := etcdClient.ConnectDB(); err != nil {
		return nil, 0, err
	}
	opts := []etcdv3.OpOption{
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByKey, etcdv3.SortAscend),
	}

	// count all match keys
	kvsAll, err := etcdClient.Kv.Get(etcdClient.Ctx, path, opts...)

	if err != nil {
		return nil, 0, err
	}

	// count total pages
	var totalPage int64
	if pageSize >= kvsAll.Count {
		totalPage = 1
	} else {
		totalPage = kvsAll.Count / pageSize
		if kvsAll.Count%pageSize > 0 {
			totalPage += 1
		}
	}

	// get start index
	pageStart := index * pageSize
	// get end index
	pageEnd := pageSize + pageStart
	if pageSize > (kvsAll.Count - pageStart) {
		pageEnd = kvsAll.Count
	}

	if pageStart >= kvsAll.Count {
		// page index do not exists set data to empty
		kvsAll.Kvs = []*mvccpb.KeyValue{}
	} else {
		// return ranged data
		kvsAll.Kvs = kvsAll.Kvs[pageStart:pageEnd]
	}

	return kvsAll, totalPage, nil
}

func (etcdClient *EtcdClient) Delete(path string) error {
	defer etcdClient.CloseConnection()
	if err := etcdClient.ConnectDB(); err != nil {
		return err
	}

	res, err := etcdClient.Kv.Delete(etcdClient.Ctx, path)
	if err != nil {
		return err
	}
	log.Debugf("Successfully delete data from etcd with revision id: %d", res.Header.Revision)
	return nil
}
