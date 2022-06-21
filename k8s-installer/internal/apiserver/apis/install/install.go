package install

import (
	"github.com/emicklei/go-restful"
	v1 "k8s-installer/internal/apiserver/apis/v1"
	"k8s-installer/pkg/server/runtime"
)

func init() {
	Install(runtime.Container)
}

func Install(c *restful.Container) {
	runtime.Must(v1.AddToContainer(c))
}
