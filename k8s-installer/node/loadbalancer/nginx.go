package loadbalancer

import "k8s-installer/schema"

type Nginx struct {
}

func (n Nginx) Start() error {
	panic("implement me")
}

func (n Nginx) Install(offline bool, task schema.ITask, resourceServerURL string) error {
	panic("implement me")
}

func (n Nginx) Remove() error {
	panic("implement me")
}

func (n Nginx) CreateConfig(listen schema.ProxySection, frontend schema.ProxySection, backend schema.ProxySection) string {
	panic("implement me")
}

func (h Nginx) CreateAPIServerConfig(address, port string) string {
	panic("implement me")
}