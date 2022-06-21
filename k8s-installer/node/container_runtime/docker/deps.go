package docker

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/schema"
)

type Deps struct {
}

func (d Deps) GetDeps() dep.DepMap {
	return dockerVersion
}

var dockerVersion dep.DepMap = schema.RegisterDep(dep.DepMap{
	"centos": {
		constants.V1_18_6: {
			"x86_64": {
				"audit-libs-python":      "audit-libs-python-2.8.5-4.el7.x86_64.rpm",
				"libcgroup":              "libcgroup-0.41-21.el7.x86_64.rpm",
				"checkpolicy":            "checkpolicy-2.5-8.el7.x86_64.rpm",
				"libsemanage-python":     "libsemanage-python-2.5-14.el7.x86_64.rpm",
				"policycoreutils":        "policycoreutils-2.5-34.el7.x86_64.rpm",
				"policycoreutils-python": "policycoreutils-python-2.5-34.el7.x86_64.rpm",
				"container-selinux":      "container-selinux-2.119.2-1.911c772.el7_8.noarch.rpm",
				"containerd.io":          "containerd.io-1.2.13-3.2.el7.x86_64.rpm",
				"docker-ce":              "docker-ce-19.03.12-3.el7.x86_64.rpm",
				"python-IPy":             "python-IPy-0.75-6.el7.noarch.rpm",
				"docker-ce-cli":          "docker-ce-cli-19.03.12-3.el7.x86_64.rpm",
				"setools-libs":           "setools-libs-3.3.8-4.el7.x86_64.rpm",
				"httpd-tools":            "httpd-tools-2.4.6-97.el7.centos.x86_64.rpm",
				"apr-util":               "apr-util-1.5.2-6.el7.x86_64.rpm",
				"apr":                    "apr-1.4.8-7.el7.x86_64.rpm",
			},
			"aarch64": {
				"containerd.io": "containerd.io-1.2.13-3.2.el7.aarch64.rpm",
				"docker-ce": "docker-ce-19.03.12-3.el7.aarch64.rpm",
				"container-selinux": "container-selinux-2.119.2-1.911c772.el7_8.noarch.rpm",
				"docker-ce-cli": "docker-ce-cli-19.03.12-3.el7.aarch64.rpm",
			},
		},
	},
	"ubuntu": {},
})
