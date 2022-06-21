package centos

import (
	"k8s-installer/node/os/family"
)

type Centos struct {
	Version family.IOSVersion
}

func (c Centos) GetOSVersion() family.IOSVersion {
	return c.Version
}

func (c Centos) GetOSFamily() string {
	return "centos"
}

func (c Centos) GetOSFullName() string {
	panic("implement me")
}
