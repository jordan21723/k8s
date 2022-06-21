package client

import (
	"k8s-installer/pkg/config/cache"
	etcdConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"

	"github.com/spf13/viper"
)

const (
	CONFIG_PATH      = "$HOME/.k8s-installer"
	CONFIG_FILE_NAME = "config-client"
	CONFIG_FILE_TYPE = "yaml"
)

type Config struct {
	ClientId              string                  `yaml:"client-id"`
	SignalPort            int                     `yaml:"signal-port"`
	Log                   log.LogConfig           `yaml:"log"`
	MessageQueue          nats.MessageQueueConfig `yaml:"message-queue"`
	Etcd                  etcdConfig.EtcdConfig   `yaml:"etcd"`
	Cache                 cache.CacheConfig       `yaml:"cache"`
	StatReportInFrequency int                     `yaml:"stat-report-in-frequency"` // unit seconds
	YamlDataDir           string                  `yaml:"yaml-data-dir"`
	Offline               bool                    `yaml:"offline"`
	ProxyIpv4CIDR         string                  `yaml:"proxy-ipv4-cidr"`
	IsTestNode            bool                    `yaml:"is-test-node"`
	CRIMountDev           string                  `yaml:"container-runtime-device"`
	LocalImagePath        string                  `yaml:"local-image-path"`
	Region                string                  `yaml:"region"`
	CRIDevMinDiskSize     uint                    `yaml:"cri-dev-min-disk-size"`
	CRIRootDir            string                  `yaml:"cri-root-dir"`
	KubeletDevMinDiskSize uint                    `yaml:"kubelet-dev-min-disk-size"`
	EtcdDevMinDiskSize    uint                    `yaml:"etcd-dev-min-disk-size"`
	IsPlexing             bool                    `yaml:"is-plexing"` // server and client process are running on the same node
	ClusterInstaller      string                  `yaml:"cluster-installer"`
}

func DefaultConfig() Config {
	return Config{
		SignalPort:            9090,
		Log:                   log.DefaultConfig(),
		MessageQueue:          nats.DefaultConfig(),
		Etcd:                  etcdConfig.DefaultConfig(),
		Cache:                 cache.DefaultConfig(),
		StatReportInFrequency: 60,
		YamlDataDir:           "/usr/share/k8s-installer/",
		Offline:               false,
		IsTestNode:            false,
		CRIMountDev:           "",
		CRIDevMinDiskSize:     150,
		KubeletDevMinDiskSize: 149,
		EtcdDevMinDiskSize:    99,
		CRIRootDir:            "/var/lib",
		ClusterInstaller:      "kubeadm",
	}
}

func DefaultConfigLookup() {
	viper.SetConfigName(CONFIG_FILE_NAME)
	viper.SetConfigType(CONFIG_FILE_TYPE)
	viper.AddConfigPath(CONFIG_PATH)
}
