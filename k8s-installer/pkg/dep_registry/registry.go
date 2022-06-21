package dep_registry

import (
	"encoding/json"
	"k8s-installer/components/velero"
	"log"
	"os"
	"reflect"
	"strings"

	"k8s-installer/components/helm"
	"k8s-installer/components/ntp"
	containerdRT "k8s-installer/node/container_runtime/containerd"
	"k8s-installer/node/container_runtime/docker"
	"k8s-installer/node/k8s"
	"k8s-installer/node/loadbalancer"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/task_breaker"
	"k8s-installer/pkg/util"
)

type InterfaceIDepRegistry interface {
	GetDeps() dep.DepMap
}

type GrafanaDeps struct {
}

func (g GrafanaDeps) GetDeps() dep.DepMap {
	return dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					// "loki":     "loki",
					// "promtail": "promtail",
					"grafana": "grafana-6.7.3-1.x86_64.rpm",
				},
				"aarch64": {
					// "loki":     "loki",
					// "promtail": "promtail",
					"grafana": "grafana-6.7.3-1.aarch64.rpm",
				},
			},
		},
		"ubuntu": {},
	}
}

type DockerPyDeps struct {
}

func (dp DockerPyDeps) GetDeps() dep.DepMap {
	return dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					"python-backports":                    "python-backports-1.0-8.el7.x86_64.rpm",
					"python-backports-ssl_match_hostname": "python-backports-ssl_match_hostname-3.5.0.1-1.el7.noarch.rpm",
					"python-ipaddress":                    "python-ipaddress-1.0.16-2.el7.noarch.rpm",
					"python-six":                          "python-six-1.9.0-2.el7.noarch.rpm",
					"python-chardet":                      "python-chardet-2.2.1-3.el7.noarch.rpm",
					"python-urllib3":                      "python-urllib3-1.10.2-7.el7.noarch.rpm",
					"python-websocket-client":             "python-websocket-client-0.56.0-3.git3c25814.el7.noarch.rpm",
					"python-requests":                     "python-requests-2.6.0-10.el7.noarch.rpm",
					"python-docker-pycreds":               "python-docker-pycreds-0.3.0-11.el7.noarch.rpm",
					"python-docker-py":                    "python-docker-py-1.10.6-9.el7_6.noarch.rpm",
				},
				"aarch64": {
					"python-backports":                    "python-backports-1.0-8.el7.aarch64.rpm",
					"python-backports-ssl_match_hostname": "python-backports-ssl_match_hostname-3.5.0.1-1.el7.noarch.rpm",
					"python-ipaddress":                    "python-ipaddress-1.0.16-2.el7.noarch.rpm",
					"python-six":                          "python-six-1.9.0-2.el7.noarch.rpm",
					"python-chardet":                      "python-chardet-2.2.1-3.el7.noarch.rpm",
					"python-urllib3":                      "python-urllib3-1.10.2-7.el7.noarch.rpm",
					"python-websocket-client":             "python-websocket-client-0.56.0-3.git3c25814.el7.noarch.rpm",
					"python-requests":                     "python-requests-2.6.0-10.el7.noarch.rpm",
					"python-docker-pycreds":               "python-docker-pycreds-0.3.0-11.el7.noarch.rpm",
					"python-docker-py":                    "python-docker-py-1.10.6-9.el7_6.noarch.rpm",
				},
			},
		},
		"ubuntu": {},
	}
}

type NFSUtilsDeps struct {
}

func (nu NFSUtilsDeps) GetDeps() dep.DepMap {
	return dep.DepMap{
		"centos": {
			constants.V1_18_6: {
				"x86_64": {
					"libcollection":     "libcollection-0.7.0-32.el7.x86_64.rpm",
					"keyutils":          "keyutils-1.5.8-3.el7.x86_64.rpm",
					"libbasicobjects":   "libbasicobjects-0.1.1-32.el7.x86_64.rpm",
					"libevent":          "libevent-2.0.21-4.el7.x86_64.rpm",
					"libpath_utils":     "libpath_utils-0.2.1-32.el7.x86_64.rpm",
					"libref_array":      "libref_array-0.1.5-32.el7.x86_64.rpm",
					"libnfsidmap":       "libnfsidmap-0.25-19.el7.x86_64.rpm",
					"libverto-libevent": "libverto-libevent-0.2.5-4.el7.x86_64.rpm",
					"libtirpc":          "libtirpc-0.2.4-0.16.el7.x86_64.rpm",
					"libnl":             "libnl-1.1.4-3.el7.x86_64.rpm",
					"quota-nls":         "quota-nls-4.01-19.el7.noarch.rpm",
					"quota-nld":         "quota-nld-4.01-19.el7.x86_64.rpm",
					"rpcbind":           "rpcbind-0.2.0-49.el7.x86_64.rpm",
					"quota":             "quota-4.01-19.el7.x86_64.rpm",
					"tcp_wrappers":      "tcp_wrappers-7.6-77.el7.x86_64.rpm",
					"gssproxy":          "gssproxy-0.7.0-29.el7.x86_64.rpm",
					"nfs-utils":         "nfs-utils-1.3.0-0.68.el7.x86_64.rpm",
				},
				"aarch64": {
					"libcollection":     "libcollection-0.7.0-32.el7.aarch64.rpm",
					"keyutils":          "keyutils-1.5.8-3.el7.aarch64.rpm",
					"libbasicobjects":   "libbasicobjects-0.1.1-32.el7.aarch64.rpm",
					"libevent":          "libevent-2.0.21-4.el7.aarch64.rpm",
					"libpath_utils":     "libpath_utils-0.2.1-32.el7.aarch64.rpm",
					"libref_array":      "libref_array-0.1.5-32.el7.aarch64.rpm",
					"libnfsidmap":       "libnfsidmap-0.25-19.el7.aarch64.rpm",
					"libverto-libevent": "libverto-libevent-0.2.5-4.el7.aarch64.rpm",
					"libtirpc":          "libtirpc-0.2.4-0.16.el7.aarch64.rpm",
					"libnl":             "libnl-1.1.4-3.el7.aarch64.rpm",
					"quota-nls":         "quota-nls-4.01-19.el7.noarch.rpm",
					"quota-nld":         "quota-nld-4.01-19.el7.aarch64.rpm",
					"rpcbind":           "rpcbind-0.2.0-49.el7.aarch64.rpm",
					"quota":             "quota-4.01-19.el7.aarch64.rpm",
					"tcp_wrappers":      "tcp_wrappers-7.6-77.el7.aarch64.rpm",
					"gssproxy":          "gssproxy-0.7.0-29.el7.aarch64.rpm",
					"nfs-utils":         "nfs-utils-1.3.0-0.68.el7.aarch64.rpm",
				},
			},
		},
		"ubuntu": {},
	}
}

var componentsLists = []InterfaceIDepRegistry{
	k8s.Deps{},
	loadbalancer.Deps{},
	containerdRT.Deps{},
	docker.Deps{},
	task_breaker.TraefikDeps{},
	task_breaker.KeepaliveDeps{},
	GrafanaDeps{},
	DockerPyDeps{},
	helm.Deps{},
	ntp.NTPDeps{},
	NFSUtilsDeps{},
	velero.Deps{},
}

func GenDepsFile() error {
	totalDeps := map[string]dep.DepMap{}
	for _, component := range componentsLists {
		totalDeps[reflect.TypeOf(component).PkgPath()+reflect.TypeOf(component).String()] = component.GetDeps()
	}
	result, err := json.MarshalIndent(totalDeps, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	currentDir, errDir := os.Getwd()
	if errDir != nil {
		log.Fatal(errDir)
	}
	if strings.Contains(currentDir, "/cmd/server") {
		// always ignore depth dir and fix dir to [path]/k8s-installer
		currentDir = strings.ReplaceAll(currentDir, "/cmd/server", "")
	}
	if err := util.WriteTxtToFileByte(currentDir+"/hack/deps/k8s-installer-deps.json", result); err != nil {
		log.Fatal(err)
	}
	return nil
}
