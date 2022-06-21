package virtual_kubelet

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

var VKMockDep dep.DepMap = schema.RegisterDep(dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"virtual-kubelet": "virtual-kubelet",
			},
			"aarch64": {
				"virtual-kubelet": "virtual-kubelet",
			},
		},
	},
	"ubuntu": {},
})
