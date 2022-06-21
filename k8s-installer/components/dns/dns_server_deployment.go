package dns

import (
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/template"
	"k8s-installer/pkg/util"
	"k8s-installer/schema/plugable"
)

type Server struct {
	Enable          bool   `json:"enable" description:"enable or disable dns server deploy,default=false"`
	ServerName      string `json:"-"`
	ServerIp        string `json:"-"`
	MQUsername      string `json:"mq_username,omitempty"  description:"message queue username if mode set to ask"`
	MQPassword      string `json:"mq_password,omitempty"  description:"message queue password if mode set to ask"`
	Upstream        string `json:"upstream,omitempty"  description:"dns upstream"`
	DataProviderURL string `json:"data_provider_url" validate:"required" description:"data provider url"`
	MQServers       string `json:"mq_servers,omitempty"  description:"message queue server address if mode set to ask, e.g. 10.0.0.200,10.0.0.202,10.0.0.203"`
	MQPort          int    `json:"mq_port,omitempty"  description:"message queue port if mode set to ask"`
	MQSubject       string `json:"mq_subject,omitempty"  description:"message queue subject if mode set to ask"`
	Mode            string `json:"mode,omitempty" validate:"oneof=ask listen" description:"dns mode, default listen"`
	AskFrequency    int    `json:"ask_frequency,omitempty"  description:"dns ask frequency if mode set to listen"`
	ImageRegistry   string `json:"-"`
}

func (server *Server) CompleteDnsServerDeploy(registry string) {
	server.MQUsername = util.StringDefaultIfNotSet(server.MQUsername, "username")
	server.MQPassword = util.StringDefaultIfNotSet(server.MQPassword, "password")
	server.MQPort = util.IntDefaultIfZero(server.MQPort, 9889)
	server.MQSubject = util.StringDefaultIfNotSet(server.MQSubject, "caas-dns")
	server.MQServers = util.StringDefaultIfNotSet(server.MQServers, "127.0.0.1")
	server.Mode = util.StringDefaultIfNotSet(server.Mode, "caas-dns")
	server.AskFrequency = util.IntDefaultIfZero(server.AskFrequency, 9889)
	server.ImageRegistry = registry
}

func (server *Server) CorefileTemplateRender() (string, error) {
	return template.New("Corefile").Render(Corefile, server)
}

func (server *Server) CaasConfigTemplateRender(serverName, serverIp string) (string, error) {
	server.ServerName = serverName
	server.ServerIp = serverIp
	return template.New("Caas Config").Render(CaasPluginConfig, server)
}

func (server *Server) DnsStaticPodTemplateRender() (string, error) {
	return template.New("Caas Config").Render(StaticPodTemplate, server)
}

func (server *Server) IsEnable() bool {
	return server.Enable
}

func (server *Server) GetName() string {
	return "DnsServerDeploy"
}

func (server *Server) GetStatus() string {
	return ""
}

func (server *Server) SetDependencies(deps map[string]plugable.IPlugAble) {

}
func (server *Server) GetDependencies() map[string]plugable.IPlugAble {
	return nil
}

func (server *Server) CheckDependencies() (error, []string) {
	return nil, nil
}

func (server *Server) GetLicenseLabel() uint16 {
	return constants.LLDnsServerDeploy
}