package cluster

import (
	"encoding/json"
	"fmt"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/schema"
	"testing"
)

var (
	clusterJSON = `
{
    "dns_server_deploy": {
        "enable": true,
        "data_provider_url": "http://172.20.149.95:8099/api/core/v1/domain/sub-domain",
        "mode": "listen",
        "mq_servers": "10.10.10.103,10.10.10.254,10.10.10.158",
        "mq_port": 9889,
        "upstream": "10.10.10.3"
    },
    "cluster_dns_upstream": {
        "enable": true,
        "addresses": [
            "10.10.10.103",
            "10.10.10.254",
            "10.10.10.158"
        ]
    },
    "ClusterId": "cluster1",
    "cluster_name": "product.core.cluster",
    "description": "core protected cluster",
    "region": "region1",
    "control_plane": {
        "service_v4_cidr": "10.96.0.0/16",
        "service_v6_cidr": "fd03::/112",
        "enable_ipvs": true
    },
    "container_runtime": {
        "container_runtime_type": "docker",
        "private_registry_port": 5000,
        "private_registry_address": "10.10.10.103",
        "private_registry_cert": "",
        "private_registry_key": "",
        "private_registry_ca": "",
        "cri_reinstall_if_already_install": false
    },
    "cni": {
        "cni_type": "calico",
        "pod_v4_cidr": "172.25.0.0/24",
        "pod_v6_cidr": "fd05::/120",
        "calico": {
            "calico_mode": "Overlay-Vxlan-All",
            "enable_dual_stack": false
        },
        "flannel": {
            "pod_v4_cidr": "",
            "tunnel_interface_ip": ""
        },
        "multus": {
            "calico": {
                "pod_v4_cidr": "",
                "ipv4_lookup_method": "",
                "pod_v6_cidr": "",
                "ipv6_lookup_method": "",
                "enable_ipip": "",
                "enable_dual_stack": false
            },
            "sriov_cni": {}
        }
    },
    "masters": [
        {
            "node-id": "192.168.234.1"
        },
        {
            "node-id": "192.168.234.2"
        },
        {
            "node-id": "192.168.234.3"
        }
    ],
    "workers": [
        {
            "node-id": "192.168.234.1"
        },
        {
            "node-id": "192.168.234.2"
        },
        {
            "node-id": "192.168.234.3"
        }
    ],
    "ingresses": {
        "enable": false,
        "ingress-nodes": [
            {
                "node-id": "192.168.234.1"
            },
            {
                "node-id": "192.168.234.2"
            },
            {
                "node-id": "192.168.234.3"
            }
        ]
    },
    "efk": {
        "enable": false,
        "namespace": "efk",
        "storage_class_name": "cinder-csi"
    },
    "pgo": {
        "enable": false,
        "storage_class": "cinder-csi"
    },
    "helm": {
        "enable": false,
        "helm_version": 3
    },
    "gap": {
        "enable": false
    },
    "middle_platform": {
        "enable": false,
        "namespace": "admin",
        "sys_admin_password": "123456",
        "db": {
            "host": "host",
            "name": "admin",
            "password": "123",
            "user": "middle",
            "replicas": 3
        }
    },
    "cloud_providers": {
        "openstack": {
            "enable": true,
            "username": "wang.qian",
            "password": "Admin,123",
            "auth_url": "https://cloud2.caas.com.cn:5000/v3",
            "project_id": "da6ad856d5a5445db222948508a306bb",
            "domain_id": "default",
            "region": "RegionOne",
            "ca_cert": "-----BEGIN CERTIFICATE-----\nMIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh\nMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3\nd3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD\nQTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT\nMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j\nb20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG\n9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB\nCSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97\nnh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt\n43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P\nT19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4\ngdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO\nBgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR\nTLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw\nDQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr\nhMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg\n06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF\nPnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls\nYSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk\nCAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIEqjCCA5KgAwIBAgIQAnmsRYvBskWr+YBTzSybsTANBgkqhkiG9w0BAQsFADBh\nMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3\nd3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD\nQTAeFw0xNzExMjcxMjQ2MTBaFw0yNzExMjcxMjQ2MTBaMG4xCzAJBgNVBAYTAlVT\nMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j\nb20xLTArBgNVBAMTJEVuY3J5cHRpb24gRXZlcnl3aGVyZSBEViBUTFMgQ0EgLSBH\nMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALPeP6wkab41dyQh6mKc\noHqt3jRIxW5MDvf9QyiOR7VfFwK656es0UFiIb74N9pRntzF1UgYzDGu3ppZVMdo\nlbxhm6dWS9OK/lFehKNT0OYI9aqk6F+U7cA6jxSC+iDBPXwdF4rs3KRyp3aQn6pj\npp1yr7IB6Y4zv72Ee/PlZ/6rK6InC6WpK0nPVOYR7n9iDuPe1E4IxUMBH/T33+3h\nyuH3dvfgiWUOUkjdpMbyxX+XNle5uEIiyBsi4IvbcTCh8ruifCIi5mDXkZrnMT8n\nwfYCV6v6kDdXkbgGRLKsR4pucbJtbKqIkUGxuZI2t7pfewKRc5nWecvDBZf3+p1M\npA8CAwEAAaOCAU8wggFLMB0GA1UdDgQWBBRVdE+yck/1YLpQ0dfmUVyaAYca1zAf\nBgNVHSMEGDAWgBQD3lA1VtFMu2bwo+IbG8OXsj3RVTAOBgNVHQ8BAf8EBAMCAYYw\nHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMBIGA1UdEwEB/wQIMAYBAf8C\nAQAwNAYIKwYBBQUHAQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdp\nY2VydC5jb20wQgYDVR0fBDswOTA3oDWgM4YxaHR0cDovL2NybDMuZGlnaWNlcnQu\nY29tL0RpZ2lDZXJ0R2xvYmFsUm9vdENBLmNybDBMBgNVHSAERTBDMDcGCWCGSAGG\n/WwBAjAqMCgGCCsGAQUFBwIBFhxodHRwczovL3d3dy5kaWdpY2VydC5jb20vQ1BT\nMAgGBmeBDAECATANBgkqhkiG9w0BAQsFAAOCAQEAK3Gp6/aGq7aBZsxf/oQ+TD/B\nSwW3AU4ETK+GQf2kFzYZkby5SFrHdPomunx2HBzViUchGoofGgg7gHW0W3MlQAXW\nM0r5LUvStcr82QDWYNPaUy4taCQmyaJ+VB+6wxHstSigOlSNF2a6vg4rgexixeiV\n4YSB03Yqp2t3TeZHM9ESfkus74nQyW7pRGezj+TC44xCagCQQOzzNmzEAP2SnCrJ\nsNE2DpRVMnL8J6xBRdjmOsC3N6cQuKuRXbzByVBjCqAA8t1L0I+9wXJerLPyErjy\nrMKWaBFLmfK/AHNF4ZihwPGOc7w6UHczBZXH5RFzJNnww+WnKuTPI0HfnVH8lg==\n-----END CERTIFICATE-----",
            "cinder": {
                "backend_type": "__DEFAULT__",
                "availability_zone": "nova",
                "storage_class_name": "cinder-csi",
                "reclaim_policy": "Delete"
            }
        }
    },
    "ClusterOperationIDs": [
        "operation1"
    ],
    "Storage": {
        "nfs": {
            "enable": false,
            "nfs_server_address": "10.0.0.200",
            "nfs_path": "/tmp/nfs_expose/",
            "reclaim_policy": "Delete",
            "storage_class_name": "nfs-sc"
        }
    },
    "console": {
        "vendor_tag": "caas-4.1",
        "tls_enable": false,
        "enable": false
    },
    "ks_cluster_conf": {
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
			"log_max_age": 7,
			"elk_prefix": "logstash"
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
			"ip_pool": "none",
			"topology": "none"
		},
		"openpitrix": {
			"enabled": true
		},
		"servicemesh": {
			"enabled": true
		}
	}
}`
	clusterMinJSON = `
{
    "dns_server_deploy": {
        "enable": true,
        "data_provider_url": "http://172.20.149.95:8099/api/core/v1/domain/sub-domain",
        "mode": "listen",
        "mq_servers": "10.10.10.103,10.10.10.254,10.10.10.158",
        "mq_port": 9889,
        "upstream": "10.10.10.3"
    },
    "cluster_dns_upstream": {
        "enable": true,
        "addresses": [
            "10.10.10.103",
            "10.10.10.254",
            "10.10.10.158"
        ]
    },
    "ClusterId": "cluster1",
    "cluster_name": "product.core.cluster",
    "description": "core protected cluster",
    "region": "region1",
    "control_plane": {
        "service_v4_cidr": "10.96.0.0/16",
        "service_v6_cidr": "fd03::/112",
        "enable_ipvs": true
    },
    "container_runtime": {
        "container_runtime_type": "docker",
        "private_registry_port": 5000,
        "private_registry_address": "10.10.10.103",
        "private_registry_cert": "",
        "private_registry_key": "",
        "private_registry_ca": "",
        "cri_reinstall_if_already_install": false
    },
    "cni": {
        "cni_type": "calico",
        "pod_v4_cidr": "172.25.0.0/24",
        "pod_v6_cidr": "fd05::/120",
        "calico": {
            "calico_mode": "Overlay-Vxlan-All",
            "enable_dual_stack": false
        },
        "flannel": {
            "pod_v4_cidr": "",
            "tunnel_interface_ip": ""
        },
        "multus": {
            "calico": {
                "pod_v4_cidr": "",
                "ipv4_lookup_method": "",
                "pod_v6_cidr": "",
                "ipv6_lookup_method": "",
                "enable_ipip": "",
                "enable_dual_stack": false
            },
            "sriov_cni": {}
        }
    },
    "masters": [
        {
            "node-id": "192.168.234.1"
        },
        {
            "node-id": "192.168.234.2"
        },
        {
            "node-id": "192.168.234.3"
        }
    ],
    "workers": [
        {
            "node-id": "192.168.234.1"
        },
        {
            "node-id": "192.168.234.2"
        },
        {
            "node-id": "192.168.234.3"
        }
    ],
    "ingresses": {
        "enable": false,
        "ingress-nodes": [
            {
                "node-id": "192.168.234.1"
            },
            {
                "node-id": "192.168.234.2"
            },
            {
                "node-id": "192.168.234.3"
            }
        ]
    },
    "efk": {
        "enable": false,
        "namespace": "efk",
        "storage_class_name": "cinder-csi"
    },
    "pgo": {
        "enable": false,
        "storage_class": "cinder-csi"
    },
    "helm": {
        "enable": false,
        "helm_version": 3
    },
    "gap": {
        "enable": false
    },
    "middle_platform": {
        "enable": false,
        "namespace": "admin",
        "sys_admin_password": "123456",
        "db": {
            "host": "host",
            "name": "admin",
            "password": "123",
            "user": "middle",
            "replicas": 3
        }
    },
    "cloud_providers": {
        "openstack": {
            "enable": true,
            "username": "wang.qian",
            "password": "Admin,123",
            "auth_url": "https://cloud2.caas.com.cn:5000/v3",
            "project_id": "da6ad856d5a5445db222948508a306bb",
            "domain_id": "default",
            "region": "RegionOne",
            "ca_cert": "-----BEGIN CERTIFICATE-----\nMIIDrzCCApegAwIBAgIQCDvgVpBCRrGhdWrJWZHHSjANBgkqhkiG9w0BAQUFADBh\nMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3\nd3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD\nQTAeFw0wNjExMTAwMDAwMDBaFw0zMTExMTAwMDAwMDBaMGExCzAJBgNVBAYTAlVT\nMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j\nb20xIDAeBgNVBAMTF0RpZ2lDZXJ0IEdsb2JhbCBSb290IENBMIIBIjANBgkqhkiG\n9w0BAQEFAAOCAQ8AMIIBCgKCAQEA4jvhEXLeqKTTo1eqUKKPC3eQyaKl7hLOllsB\nCSDMAZOnTjC3U/dDxGkAV53ijSLdhwZAAIEJzs4bg7/fzTtxRuLWZscFs3YnFo97\nnh6Vfe63SKMI2tavegw5BmV/Sl0fvBf4q77uKNd0f3p4mVmFaG5cIzJLv07A6Fpt\n43C/dxC//AH2hdmoRBBYMql1GNXRor5H4idq9Joz+EkIYIvUX7Q6hL+hqkpMfT7P\nT19sdl6gSzeRntwi5m3OFBqOasv+zbMUZBfHWymeMr/y7vrTC0LUq7dBMtoM1O/4\ngdW7jVg/tRvoSSiicNoxBN33shbyTApOB6jtSj1etX+jkMOvJwIDAQABo2MwYTAO\nBgNVHQ8BAf8EBAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUA95QNVbR\nTLtm8KPiGxvDl7I90VUwHwYDVR0jBBgwFoAUA95QNVbRTLtm8KPiGxvDl7I90VUw\nDQYJKoZIhvcNAQEFBQADggEBAMucN6pIExIK+t1EnE9SsPTfrgT1eXkIoyQY/Esr\nhMAtudXH/vTBH1jLuG2cenTnmCmrEbXjcKChzUyImZOMkXDiqw8cvpOp/2PV5Adg\n06O/nVsJ8dWO41P0jmP6P6fbtGbfYmbW0W5BjfIttep3Sp+dWOIrWcBAI+0tKIJF\nPnlUkiaY4IBIqDfv8NZ5YBberOgOzW6sRBc4L0na4UU+Krk2U886UAb3LujEV0ls\nYSEY1QSteDwsOoBrp+uvFRTp2InBuThs4pFsiv9kuXclVzDAGySj4dzp30d8tbQk\nCAUw7C29C79Fv1C5qfPrmAESrciIxpg0X40KPMbp1ZWVbd4=\n-----END CERTIFICATE-----\n-----BEGIN CERTIFICATE-----\nMIIEqjCCA5KgAwIBAgIQAnmsRYvBskWr+YBTzSybsTANBgkqhkiG9w0BAQsFADBh\nMQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3\nd3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD\nQTAeFw0xNzExMjcxMjQ2MTBaFw0yNzExMjcxMjQ2MTBaMG4xCzAJBgNVBAYTAlVT\nMRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxGTAXBgNVBAsTEHd3dy5kaWdpY2VydC5j\nb20xLTArBgNVBAMTJEVuY3J5cHRpb24gRXZlcnl3aGVyZSBEViBUTFMgQ0EgLSBH\nMTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBALPeP6wkab41dyQh6mKc\noHqt3jRIxW5MDvf9QyiOR7VfFwK656es0UFiIb74N9pRntzF1UgYzDGu3ppZVMdo\nlbxhm6dWS9OK/lFehKNT0OYI9aqk6F+U7cA6jxSC+iDBPXwdF4rs3KRyp3aQn6pj\npp1yr7IB6Y4zv72Ee/PlZ/6rK6InC6WpK0nPVOYR7n9iDuPe1E4IxUMBH/T33+3h\nyuH3dvfgiWUOUkjdpMbyxX+XNle5uEIiyBsi4IvbcTCh8ruifCIi5mDXkZrnMT8n\nwfYCV6v6kDdXkbgGRLKsR4pucbJtbKqIkUGxuZI2t7pfewKRc5nWecvDBZf3+p1M\npA8CAwEAAaOCAU8wggFLMB0GA1UdDgQWBBRVdE+yck/1YLpQ0dfmUVyaAYca1zAf\nBgNVHSMEGDAWgBQD3lA1VtFMu2bwo+IbG8OXsj3RVTAOBgNVHQ8BAf8EBAMCAYYw\nHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMBIGA1UdEwEB/wQIMAYBAf8C\nAQAwNAYIKwYBBQUHAQEEKDAmMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdp\nY2VydC5jb20wQgYDVR0fBDswOTA3oDWgM4YxaHR0cDovL2NybDMuZGlnaWNlcnQu\nY29tL0RpZ2lDZXJ0R2xvYmFsUm9vdENBLmNybDBMBgNVHSAERTBDMDcGCWCGSAGG\n/WwBAjAqMCgGCCsGAQUFBwIBFhxodHRwczovL3d3dy5kaWdpY2VydC5jb20vQ1BT\nMAgGBmeBDAECATANBgkqhkiG9w0BAQsFAAOCAQEAK3Gp6/aGq7aBZsxf/oQ+TD/B\nSwW3AU4ETK+GQf2kFzYZkby5SFrHdPomunx2HBzViUchGoofGgg7gHW0W3MlQAXW\nM0r5LUvStcr82QDWYNPaUy4taCQmyaJ+VB+6wxHstSigOlSNF2a6vg4rgexixeiV\n4YSB03Yqp2t3TeZHM9ESfkus74nQyW7pRGezj+TC44xCagCQQOzzNmzEAP2SnCrJ\nsNE2DpRVMnL8J6xBRdjmOsC3N6cQuKuRXbzByVBjCqAA8t1L0I+9wXJerLPyErjy\nrMKWaBFLmfK/AHNF4ZihwPGOc7w6UHczBZXH5RFzJNnww+WnKuTPI0HfnVH8lg==\n-----END CERTIFICATE-----",
            "cinder": {
                "backend_type": "__DEFAULT__",
                "availability_zone": "nova",
                "storage_class_name": "cinder-csi",
                "reclaim_policy": "Delete"
            }
        }
    },
    "ClusterOperationIDs": [
        "operation1"
    ],
    "Storage": {
        "nfs": {
            "enable": false,
            "nfs_server_address": "10.0.0.200",
            "nfs_path": "/tmp/nfs_expose/",
            "reclaim_policy": "Delete",
            "storage_class_name": "nfs-sc"
        }
    },
    "console": {
        "vendor_tag": "caas-4.1",
        "tls_enable": false,
        "enable": false
    },
    "ks_cluster_conf": {
		"enabled": true,
		"is_managed_by_kubefed": true,
		"member_of_cluster": "cluster-id",
		"server": {
			"local_registry_server": "127.0.0.1:5000",
			"storage_class": "nfs"
		},
		"monitor": {
		},
		"etcd_monitor": {
		},
		"es": {
		},
		"console": {
			"enable_multi_login": true,
			"port": 30880
		},
		"alerting": {
			"enabled": true
		},
		"auditing": {
			"enabled": true
		},
		"devops": {
			"enabled": true
		},
		"events": {
			"enabled": true
		},
		"logging": {
			"enabled": true
		},
		"metrics_server": {
			"enabled": true
		},
		"multicluster": {
			"cluster_role": "host",
			"cluster_type": "production"
		},
		"network": {
			"np_enabled": true
		},
		"openpitrix": {
			"enabled": true
		},
		"servicemesh": {
			"enabled": true
		}
	}
}`
)

func TestCreateCluster(t *testing.T) {
	c := &schema.Cluster{}
	if err := json.Unmarshal([]byte(clusterMinJSON), c); err != nil {
		t.Error(err)
	}
	if err := utils.Validate(c); err != nil {
		t.Error(err)
	}
	fmt.Printf("%#v\n", c)
	c.KsClusterConf = c.KsClusterConf.CompleteDeploy()
	fmt.Printf("%#v\n", c.KsClusterConf)
	result, err := c.KsClusterConf.TemplateRender()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result)
}
