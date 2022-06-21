package schema

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/schema/plugable"
)

type Harbor struct {
	Enable    bool   `json:"enable,omitempty" description:"enable or disable harbor interfacing,default=false"`
	Host      string `json:"host,omitempty" validate:"omitempty,fqdn" optional:"true" description:"harbor address e.g. caas.registry.com or caas.registry.com:8038"`
	Ip        string `json:"ip" validate:"ipv4|ipv6" description:"ip of host resolve"`
	Port      int    `json:"port" validate:"omitempty,max=65535,min=1" description:"private registry port e.g. 4000"`
	UserName  string `json:"user_name" description:"harbor username"`
	Password  string `json:"password" description:"password of harbor username "`
	EnableTls bool   `json:"enable_tls,omitempty" description:"whether enable tls or not"`
	Cert      string `json:"cert,omitempty" description:"cert string"`
	Key       string `json:"key,omitempty" description:"key string"`
	CA        string `json:"ca,omitempty" description:"ca string"`
}

func (h Harbor) IsEnable() bool {
	return h.Enable
}

func (h Harbor) GetName() string {
	return "Harbor"
}

func (h Harbor) GetStatus() string {
	return ""
}

func (h Harbor) SetDependencies(map[string]plugable.IPlugAble) {
}

func (h Harbor) GetDependencies() map[string]plugable.IPlugAble {
	return map[string]plugable.IPlugAble{}
}

func (h Harbor) CheckDependencies() (error, []string) {
	return nil,nil
}

func (h Harbor) GetLicenseLabel() uint16 {
	return constants.LLHarbor
}
