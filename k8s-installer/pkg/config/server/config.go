package server

import (
	apiServerConfig "k8s-installer/pkg/config/api_server"
	"k8s-installer/pkg/config/cache"
	etcdConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/coredns"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"

	"github.com/spf13/viper"
)

const (
	CONFIG_PATH      = "$HOME/.k8s-installer"
	CONFIG_ETC_PATH  = "/etc/k8s-installer"
	CONFIG_FILE_NAME = "config-server"
	CONFIG_FILE_TYPE = "yaml"
)

type Config struct {
	ServerId         string                    `yaml:"server-id"`
	Log              log.LogConfig             `yaml:"log"`
	MessageQueue     nats.MessageQueueConfig   `yaml:"message-queue"`
	ApiServer        apiServerConfig.ApiConfig `yaml:"api-server"`
	Cache            cache.CacheConfig         `yaml:"cache"`
	Etcd             etcdConfig.EtcdConfig     `yaml:"etcd"`
	SignalPort       int                       `yaml:"signal-port"`
	HostRequirement  HostConfigRequireConfig   `yaml:"host-requirement"`
	TaskTimeOut      SeverMessageTimeOutConfig `yaml:"server-message-timeout-config"`
	SystemInfoPubKey string                    `yaml:"system-info-pubkey"`
	/*
		lazy operation logs meaning if no error or time occur, programme will not save operation logs to db immediately
		instead logs will save to db either error or all done
	*/
	DisableLazyOperationLog        bool `yaml:"disable-lazy-operation-log"`
	EnableAgentLivenessDetector    bool `yaml:"enable-agent-liveness-detector-frequency"`
	AgentLivenessDetectorFrequency int  `yaml:"agent-liveness-detector-frequency"`
	IsTestNode                     bool `yaml:"is-test-node"`
	// deprecated ReportToKubeCaas
	ReportToKubeCaas        bool           `yaml:"report-to-kube-caas"`
	EnableHostnameRename    bool           `yaml:"enable-hostname-rename"`
	ConsoleTitle            string         `yaml:"console-title"`
	ConsoleResourcePath     string         `yaml:"console-resource-path"`
	RegistryDevMinDiskSize  uint           `yaml:"cri-dev-min-disk-size"`
	DockerDevMinDiskSize    uint           `yaml:"docker-dev-min-disk-size"`
	EtcdDevMinDiskSize      uint           `yaml:"etcd-dev-min-disk-size"`
	K8sEtcdDataDir          string         `yaml:"k8s-etcd-data-dir"`
	CoreDNSConfig           coredns.Config `yaml:"coredns-config"`
	BackupPath              string         `yaml:"backup-path"`
	MaxContainerLogFileSize int32          `yaml:"max_container_log_size"`
	MaxContainerLogFile     int32          `yaml:"max_container_log_file"`
	MesssagePlatformHost    string         `yaml:"messsage_platform_host"`
	Mock                    bool           `yaml:"mock"`
}

func DefaultConfig() Config {
	return Config{
		MaxContainerLogFile:     3,
		MaxContainerLogFileSize: 10,
		CoreDNSConfig:           coredns.DefaultConfig(),
		RegistryDevMinDiskSize:  249,
		DockerDevMinDiskSize:    249,
		EtcdDevMinDiskSize:      149,
		Log:                     log.DefaultConfig(),
		MessageQueue:            nats.DefaultConfig(),
		K8sEtcdDataDir:          "/var/lib",
		ApiServer: apiServerConfig.ApiConfig{
			ApiVipAddress:          "127.0.0.1:8099",
			ApiPort:                8099,
			ResourceServerPort:     8079,
			ResourceServerFilePath: "/etc/k8s-installer/resource",
			JWTExpiredAfterHours:   8,
			JWTSignMethod:          "HS256",
			JWTSignString: `LS0tLS1CRUdJTiBPUEVOU1NIIFBSSVZBVEUgS0VZLS0tLS0KYjNCbGJuTnphQzFyWlhrdGRqRUFB
QUFBQkc1dmJtVUFBQUFFYm05dVpRQUFBQUFBQUFBQkFBQUJGd0FBQUFkemMyZ3RjbgpOaEFBQUFB
d0VBQVFBQUFRRUFyYkdsenMrZHUwZmI4NE5IbU8wa1VtRU9iZEFkSisvaGhsZ3QzTFU0WWxBUXRO
cUY4VUw3Clp5VEY3cDdTcFB0UnFRY1FyUTJzbE8zQmEydEttZFRhcUdIOWF2YVpWRm5sOXU4Vmc4
NG02NGdoTjhuTlI4QnN0QjFBTjkKbDl1K1BQYTR5TFF5WkY1dzZaSVA0Vy9RV2tkaHRxYWE5WkZR
Sk5qeW55N3FZc1B5QTRPODI4dFNOaldRSHU2YUNVK1RiTgpLRU1sUnB2RmhWMStmOTJxWGpkalBi
YS9DTys0bUhNNTMvby9VeW5TSlZ2eU9WUE4vdGErN29GQUhydnFLZXFYbFR5eEx0CnZveGNVWHpv
ZzNlY1F1NGdRaFFLSzAzSWZ4aG9LSjNaYnA0VGtxdm9HUGJ0S2NmYStIZ0Z1eXFncTI3T1dKYSta
Q2oweEMKclFLQUtpQWo3UUFBQThnOFNSNmtQRWtlcEFBQUFBZHpjMmd0Y25OaEFBQUJBUUN0c2FY
T3o1MjdSOXZ6ZzBlWTdTUlNZUQo1dDBCMG43K0dHV0MzY3RUaGlVQkMwMm9YeFF2dG5KTVh1bnRL
aysxR3BCeEN0RGF5VTdjRnJhMHFaMU5xb1lmMXE5cGxVCldlWDI3eFdEemlicmlDRTN5YzFId0d5
MEhVQTMyWDI3NDg5cmpJdERKa1huRHBrZy9oYjlCYVIyRzJwcHIxa1ZBazJQS2YKTHVwaXcvSURn
N3pieTFJMk5aQWU3cG9KVDVOczBvUXlWR204V0ZYWDUvM2FwZU4yTTl0cjhJNzdpWWN6bmYrajlU
S2RJbApXL0k1VTgzKzFyN3VnVUFldStvcDZwZVZQTEV1MitqRnhSZk9pRGQ1eEM3aUJDRkFvclRj
aC9HR2dvbmRsdW5oT1NxK2dZCjl1MHB4OXI0ZUFXN0txQ3JiczVZbHI1a0tQVEVLdEFvQXFJQ1B0
QUFBQUF3RUFBUUFBQVFBK1F1T3drbk56NG5wUmU4bDYKWStjVk1IMC9sODRidHIwY3J4Y2hla1JQ
Mld0anFNRkNqa1FYNFBLaWFvUVBaNWNLQStKU1pnaHJDaDYvSnFLREtlMkhWagpqRTBzaDdtQTM2
eWhEb1FrbHBQRTdMOUthRkJkRHhiMXJKcWtpTHhVbGd2K3hia2FpVS9vS2RkUGRBazNrMGJQZGtF
dHJYCjBRK0VOZ0ZDMG9ZaHlnOVZJSDVFQVdHOEgxZWRaOTVRRGsyZmJaYmJkdmJKUExJRjJEeWNr
ZlBLUmpZRkZKa2paV3hzcDMKUzJ4cXpkaE1IblNvZ0RWZ0ptalhKQ00vUTZZdTdBU1NvVTN0OFNl
VThtY1J0ZitDZUM1VnNNbkt3MFZNdjNmNnlCcmhObwprWGNsR0p0NVZrT3Y5ZlpXc3ZmUHprRFY4
M0RrOUhYVWNwZS9SNlYwVWJCeEFBQUFnQVA1QkNkbURKZ28vYkxONTZuZUJnCnoxNkpSdnBCbGFX
aENNTmZFNWUzVVZnQkpPWXRQRnpEN3NrSWZOVzRFdXQ3Rm94clJRUjBpenhHcGFuUTNRNitmNHR3
MW0KVi9ScGY5ZGdaV0dpS3dsR2gwMHlDQkx3Y2dPaXl5cUpNbllIaWJ1U1ZXOU1KanFSaE53T05C
NGtWczhhVUxEVWZVNWkvLwpKS2FhTkVrZmFMQUFBQWdRRGg3WlZUVjBHQ0lxK0Uwd24yckJwNlhx
RzA4dTZMTVpwbURZMFFXR3BvZElaWjEvazBmbzRFCngxN1RNZ3ZqNlkvYlBWZkl3TXVCS0ovNk5W
OVFsUmJydm5GYTd4elVGUENoWFRZNTVNNDFGaTJ0VDJNS2N0STA4VHlDNk8KSlFRYUJCNTlURXJx
NlRQclZWLzVySFptSVB2TEtKY1hibTNIUHlybmpLMGxGbmt3QUFBSUVBeE5BMG9EMS8xQk5PMTh0
YgozUENiZjZwdDVDeitQRitxNlZlb082U2N1T3VEL3hHNWQ0RC83ZnEweHNWVEVlTTB3UWFTYVVK
c3BQeC93VVhyMmZzWXRFCjNMUkh4cGp5bDdJV2NTOE14cFhNRWE5RDl6Y04wMTdmZHd5Y2tMWCtI
akdnWGlpQW1vTUE5UTlHSXU2VHFkQWZ1S1VtYTkKV2tzNmhGNUVRem9BZG44QUFBQU9jbTl2ZEVC
emFXMXZibk10Y0dNQkFnTUVCUT09Ci0tLS0tRU5EIE9QRU5TU0ggUFJJVkFURSBLRVktLS0tLQo=`,
		},
		HostRequirement: HostConfigRequireConfig{
			CheckKernelVersion:     false,
			CheckOSFamily:          true,
			KernelMajorVersion:     3,
			KernelSubVersion:       10,
			KernelTailVersion:      0,
			SupportOSFamily:        "centos,ubuntu",
			SupportOSFamilyVersion: "7,",
		},
		Cache:      cache.DefaultConfig(),
		Etcd:       etcdConfig.DefaultConfig(),
		SignalPort: 9090,
		TaskTimeOut: SeverMessageTimeOutConfig{
			TaskBasicConfig:                  5,
			TaskKubectl:                      5,
			TaskCRI:                          300,
			TaskKubeadmInitFirstControlPlane: 300,
			TaskKubeadmJoinControlPlane:      300,
			TaskKubeadmJoinWorker:            300,
			TaskVip:                          60,
			TaskRenameHostName:               5,
		},
		AgentLivenessDetectorFrequency: 110,
		EnableAgentLivenessDetector:    false,
		IsTestNode:                     false,
		BackupPath:                     "/usr/share/minio/data",
	}
}

type SeverMessageTimeOutConfig struct {
	TaskPrepareOfflineResource       int `yaml:"task-prepare-offline-resource"`
	TaskBasicConfig                  int `yaml:"task-basic-config-timeout"`
	TaskKubectl                      int `yaml:"task-kubectl-timeout"`
	TaskCRI                          int `yaml:"task-cri-timeout"`
	TaskKubeadmInitFirstControlPlane int `yaml:"task-kubeadm-init-first-control-plane-timeout"`
	TaskKubeadmJoinControlPlane      int `yaml:"task-kubeadm-join-control-plane-timeout"`
	TaskKubeadmJoinWorker            int `yaml:"task-kubeadm-join-worker-timeout"`
	TaskVip                          int `yaml:"task-vip-timeout"`
	TaskRenameHostName               int `yaml:"task-rename-hostname-timeout"`
}

type HostConfigRequireConfig struct {
	CheckKernelVersion     bool   `yaml:"check-kernel-version"`
	CheckOSFamily          bool   `yaml:"check-os-family"`
	KernelMajorVersion     int    `yaml:"kernel-major"`
	KernelSubVersion       int    `yaml:"kernel-sub"`
	KernelTailVersion      int    `yaml:"kernel-ver"`
	SupportOSFamily        string `yaml:"support-os-family"`
	SupportOSFamilyVersion string `yaml:"support-os-family-version"`
}

func DefaultConfigLookup() {
	viper.SetConfigName(CONFIG_FILE_NAME)
	viper.SetConfigType(CONFIG_FILE_TYPE)
	viper.AddConfigPath(CONFIG_PATH)
	viper.AddConfigPath(CONFIG_ETC_PATH)
}
