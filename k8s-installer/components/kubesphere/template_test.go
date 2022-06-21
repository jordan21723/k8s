package kubesphere

import (
	"encoding/json"
	"fmt"
	"testing"
)

var (
	KSClusterConfigRawJSON = `
{
	"enabled": true,
	"is_managed_by_kubefed": true,
	"member_of_cluster": "cluster-id",
	"server": {
		"local_registry_server": "127.0.0.1:5000",
		"storage_class": "nfs",
		"jwt_secret": "caas",
		"minio_volume_size": "20Gi",
		"openldap_volume_size": "20Gi",
		"redis_volume_size": "20Gi"
	},
	"monitor": {
		"storage_class": "cinder",
		"prometheus_replicas": 1,
		"prometheus_memory_request": "400Mi",
		"prometheus_volume_size": "50Gi",
		"alertmanager_replicas": 1,
		"monitor_endpoint": "http://prometheus-operated.kubesphere-monitoring-system.svc:9090"
	},
	"etcd_monitor": {
		"enabled": false,
		"endpoint_ips": ["127.0.0.1"],
		"port": 2379,
		"tls_enable": false
	},
	"es": {
		"elasticsearch_master_replicas": 1,
		"elasticsearch_data_replicas": 2,
		"elasticsearch_master_volume_size": "20Gi",
		"elasticsearch_data_volume_size": "40Gi",
		"log_max_age": 9,
		"elk_prefix": "logstash2"
	},
	"console": {
		"enable_multi_login": true,
		"port": 30880
	},
	"alerting": {
		"enabled": true,
		"ruler_replica": 1
	},
	"auditing": {
		"enabled": true
	},
	"devops": {
		"enabled": true,
		"jenkins_memory_limit": "2Gi",
		"jenkins_memory_request": "1500Mi",
		"jenkins_volume_size": "50Gi",
		"jenkins_java_opts_xms": "512m",
		"jenkins_java_opts_xmx": "512m",
		"jenkins_java_opts_max_ram": "2g"
	},
	"events": {
		"enabled": true,
		"ruler_replicas": 1
	},
	"logging": {
		"enabled": true,
		"logsidecar_replicas": 1
	},
	"metrics_server": {
		"enabled": true
	},
	"multicluster": {
		"cluster_role": "host",
		"cluster_type": "production"
	},
	"network": {
		"np_enabled": true,
		"ip_pool": "calico",
		"topology": "weave-scope"
	},
	"openpitrix": {
		"enabled": true
	},
	"servicemesh": {
		"enabled": true
	}
}
`
)

func TestKSTemplate31Render(t *testing.T) {
	c := &KSClusterConfig{}
	if err := json.Unmarshal([]byte(KSClusterConfigRawJSON), c); err != nil {
		t.Error(err)
	}
	result, err := c.TemplateRender()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)
}
