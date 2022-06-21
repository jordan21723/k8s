package loadbalancer

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type Deps struct {
}

func (k Deps) GetDeps() dep.DepMap {
	return HaproxyVersionMapping
}

var HaproxyVersionMapping dep.DepMap = schema.RegisterDep(dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"haproxy": "haproxy-1.5.18-9.el7.x86_64.rpm",
			},
			"aarch64": {
				"haproxy": "haproxy-1.5.18-9.el7.aarch64.rpm",
			},
		},
	},
	"ubuntu": {},
})
