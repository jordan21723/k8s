module k8s-installer

go 1.15

require (
	github.com/containerd/cgroups v1.0.1 // indirect
	github.com/containerd/containerd v1.4.4
	github.com/containerd/continuity v0.1.0 // indirect
	github.com/containerd/fifo v1.0.0 // indirect
	github.com/containerd/go-runc v1.0.0 // indirect
	github.com/containerd/ttrpc v1.0.2 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/emicklei/go-restful v2.13.0+incompatible
	github.com/emicklei/go-restful-openapi v1.4.1
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/go-openapi/spec v0.19.3
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.0
	github.com/hyperboloide/lk v0.0.0-20200504060759-b535f1973118
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.3 // indirect
	github.com/mitchellh/mapstructure v1.3.0
	github.com/nats-io/nats-server/v2 v2.4.0
	github.com/nats-io/nats.go v1.12.0
	github.com/shirou/gopsutil v0.0.0-20180427012116-c95755e4bcd7
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/thoas/go-funk v0.9.0
	github.com/txn2/txeh v1.3.0
	github.com/zcalusic/sysinfo v0.0.0-20210905121133-6fa2f969a900
	go.etcd.io/etcd v0.0.0-20201125193152-8a03d2e9614b
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools/v3 v3.0.3 // indirect
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	k8s.io/client-go v0.18.6
	k8s.io/kube-proxy v0.18.6
	k8s.io/kubelet v0.18.6
	k8s.io/kubernetes v1.18.6
)

replace (
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/apiserver => k8s.io/apiserver v0.18.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.6
	k8s.io/code-generator => k8s.io/code-generator v0.18.6
	k8s.io/component-base => k8s.io/component-base v0.18.6
	k8s.io/cri-api => k8s.io/cri-api v0.18.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.6
	k8s.io/kubectl => k8s.io/kubectl v0.18.6
	k8s.io/kubelet => k8s.io/kubelet v0.18.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.6
	k8s.io/metrics => k8s.io/metrics v0.18.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.6
)
