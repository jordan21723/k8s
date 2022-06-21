package ubuntu

import "k8s-installer/node/os/family"

type Ubuntu struct {
	Version family.IOSVersion
}

func (c Ubuntu) GetOSVersion() family.IOSVersion {
	return c.Version
}

func (c Ubuntu) GetOSFamily() string {
	return "ubuntu"
}

func (c Ubuntu) GetOSFullName() string {
	panic("implement me")
}