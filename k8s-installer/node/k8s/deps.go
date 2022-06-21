package k8s

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type Deps struct {
}

func (k Deps) GetDeps() dep.DepMap {
	return KubeDepMapping
}

var KubeDepMapping dep.DepMap = schema.RegisterDep(dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"conntrack":             "conntrack-tools-1.4.4-7.el7.x86_64.rpm",
				"kubeadm":               "kubeadm-1.18.6-0.x86_64.rpm",
				"libnetfilter_cthelper": "libnetfilter_cthelper-1.0.0-11.el7.x86_64.rpm",
				// "libsemanage-python":     "libsemanage-python-2.5-14.el7.x86_64.rpm",
				"cri-tools":              "cri-tools-1.13.0-0.x86_64.rpm",
				"kubectl":                "kubectl-1.18.6-0.x86_64.rpm",
				"libnetfilter_cttimeout": "libnetfilter_cttimeout-1.0.0-7.el7.x86_64.rpm",
				"ebtables":               "ebtables-2.0.10-16.el7.x86_64.rpm",
				"kubelet":                "kubelet-1.18.6-0.x86_64.rpm",
				"libnetfilter_queue":     "libnetfilter_queue-1.0.2-2.el7_2.x86_64.rpm",
				"ipvsadm":                "ipvsadm-1.27-8.el7.x86_64.rpm",
				"kubernetes-cni":         "kubernetes-cni-0.8.6-0.x86_64.rpm",
				"socat":                  "socat-1.7.3.2-2.el7.x86_64.rpm",
			},
			"aarch64": {
				"conntrack":              "conntrack-tools-1.4.4-7.el7.aarch64.rpm",
				"kubeadm":                "kubeadm-1.18.6-0.aarch64.rpm",
				"libnetfilter_cthelper":  "libnetfilter_cthelper-1.0.0-11.el7.aarch64.rpm",
				"cri-tools":              "cri-tools-1.13.0-0.aarch64.rpm",
				"kubectl":                "kubectl-1.18.6-0.aarch64.rpm",
				"libnetfilter_cttimeout": "libnetfilter_cttimeout-1.0.0-7.el7.aarch64.rpm",
				"ebtables":               "ebtables-2.0.10-16.el7.aarch64.rpm",
				"kubelet":                "kubelet-1.18.6-0.aarch64.rpm",
				"libnetfilter_queue":     "libnetfilter_queue-1.0.2-2.el7.aarch64.rpm",
				"ipvsadm":                "ipvsadm-1.27-8.el7.aarch64.rpm",
				"kubernetes-cni":         "kubernetes-cni-0.8.7-0.aarch64.rpm",
				"socat":                  "socat-1.7.3.2-2.el7.aarch64.rpm",
			},
		},
	},
	"ubuntu": {},
})
