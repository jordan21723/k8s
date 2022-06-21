package dns

import (
	"fmt"
	"k8s-installer/pkg/template"
	"k8s-installer/schema/plugable"
)

type UpStreamServer struct {
	Enable          bool     `json:"enable" description:"enable or disable dns server deploy,default=false"`
	Addresses       []string `json:"addresses"`
	ClusterName     string   `json:"-"`
	CombinedAddress string   `json:"-"`
}

func (upStr *UpStreamServer) IsEnable() bool {
	return upStr.Enable
}

func (upStr *UpStreamServer) GetName() string {
	return "UpStreamServer"
}

func (upStr *UpStreamServer) GetStatus() string {
	return ""
}

func (upStr *UpStreamServer) SetDependencies(deps map[string]plugable.IPlugAble) {

}
func (upStr *UpStreamServer) GetDependencies() map[string]plugable.IPlugAble {
	return nil
}

func (upStr *UpStreamServer) CheckDependencies() (error, []string) {
	return nil, nil
}

func (upStr *UpStreamServer) GetLicenseLabel() uint16 {
	return 0
}

func (upStr *UpStreamServer) ClusterDnsTemplateRender(clusterName string) (string, error) {
	if len(upStr.Addresses) > 0 {
		for _, addr := range upStr.Addresses {
			upStr.CombinedAddress += fmt.Sprintf("%s ", addr)
		}
	} else {
		upStr.CombinedAddress = "/etc/resolv.conf"
	}
	upStr.ClusterName = clusterName
	return template.New("ClusterDnsTemplate").Render(ClusterCorednsConfig, upStr)
}