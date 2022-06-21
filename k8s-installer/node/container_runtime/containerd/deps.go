package containerd

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type Deps struct {
}

func (C Deps) GetDeps() dep.DepMap {
	return containerdVersion
}

var containerdVersion dep.DepMap = schema.RegisterDep(dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"device-mapper-persistent-data": "device-mapper-persistent-data-0.8.5-3.el7_9.2.x86_64.rpm",
				"lvm2-2.02.186-7":               "lvm2-2.02.186-7.el7_8.2.src.rpm",
				//"containerd.io":                 "containerd.io-1.2.13-3.2.el7.x86_64.rpm",
				"containerd.io": "containerd.io-1.4.4-3.1.el7.x86_64.rpm",
				"kata-runtime":  "kata-runtime-1.11.5-10.1.x86_64.rpm",
				"kata-proxy":    "kata-proxy-1.11.5-10.1.x86_64.rpm",
				"kata-shim":     "kata-shim-1.11.5-10.1.x86_64.rpm",
				"httpd-tools":   "httpd-tools-2.4.6-97.el7.centos.x86_64.rpm",
				"apr-util":      "apr-util-1.5.2-6.el7.x86_64.rpm",
				"apr":           "apr-1.4.8-7.el7.x86_64.rpm",
				"mock":          "mock-2.6-1.el7.noarch.rpm",
				"mock-core":     "mock-core-configs-33.3-1.el7.noarch.rpm",
			},
			"aarch64": {
				"device-mapper-persistent-data": "device-mapper-persistent-data-0.8.5-2.el7.aarch64.rpm",
				"7:lvm2-2.02.186-7":             "7:lvm2-2.02.186-7.el7_8.2.aarch64.rpm",
				//"containerd.io":                 "containerd.io-1.2.13-3.2.el7.aarch64.rpm",
				"containerd.io": "containerd.io-1.4.4-3.1.el7.aarch64.rpm",
				"kata-runtime":  "kata-runtime-1.12.0~alpha0-68.1.aarch64.rpm",
				"kata-proxy":    "kata-proxy-1.12.0~alpha0-48.1.aarch64.rpm",
				"kata-shim":     "kata-shim-1.12.0~alpha0-46.1.aarch64.rpm",
				"httpd-tools":   "httpd-tools-2.4.6-95.el7.centos.aarch64.rpm",
				"apr-util":      "apr-util-1.5.2-6.el7.x86_64.rpm",
			},
		},
	},
	"ubuntu": {},
})
