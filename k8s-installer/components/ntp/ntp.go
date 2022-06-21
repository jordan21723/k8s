package ntp

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
)

type NTPDeps struct {
}

func (g NTPDeps) GetDeps() dep.DepMap {
	return dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					"ntp": "chrony-3.4-1.el7.x86_64.rpm",
				},
				"aarch64": {
					"ntp": "chrony-3.4-1.el7.aarch64.rpm",
				},
			},
		},
		"ubuntu": {},
	}
}
