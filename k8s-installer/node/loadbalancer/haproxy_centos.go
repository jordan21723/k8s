package loadbalancer

import (
	"bytes"
	"fmt"
	"strings"

	"k8s-installer/node/os/family"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
)

type Haproxy struct {
}

func (h Haproxy) Install(offline bool, task schema.TaskLoadBalance, cluster schema.Cluster, config client.Config, resourceServerURL, osVersion, cpuArch string, md5Dep dep.DepMap) error {
	var stdErr bytes.Buffer
	var err error

	// install yum-utils
	if offline {
		log.Debug("Installing haproxy offline")
		saveTo := config.YamlDataDir + "/haproxy"
		if err := family.CommonDownloadDep(resourceServerURL, HaproxyVersionMapping, saveTo, constants.V1_18_6, md5Dep); err != nil {
			return err
		}

		var stdErr bytes.Buffer
		var err error
		// install yum-utils
		log.Debug("Installing all haproxy rpms")
		_, stdErr, err = command.RunCmd("rpm", "-ivh", "--replacefiles", "--replacepkgs", "--nodeps", saveTo+"/*.rpm")
		if err != nil {
			log.Errorf("Failed to install haproxy related rpm due to following error:")
			log.Errorf("StdErr %s", stdErr.String())
			return err
		}
	} else {
		log.Debug("Installing haproxy online")
		_, stdErr, err = command.RunCmd("yum", "install", "haproxy", "-y")
		if err != nil {
			log.Errorf("Failed to install haproxy due to following error:")
			log.Errorf("StdErr %s", stdErr.String())
			return err
		}
	}

	log.Debug("Creating haproxy config to /etc/haproxy/haproxy.cfg")

	if err := util.WriteTxtToFile("/etc/haproxy/haproxy.cfg", string(task.ProxyConfig)); err != nil {
		return err
	}
	return err
}

func (h Haproxy) Remove() error {
	log.Debug("Try to remove haproxy")
	var stdErr bytes.Buffer
	var err error
	_, stdErr, err = command.RunCmd("yum", "remove", "haproxy", "-y")
	if err != nil {
		log.Errorf("Failed to remove haproxy due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	return nil
}

func (h Haproxy) CreateConfig(sections []schema.ProxySection) string {
	appendSection := ""
	for _, section := range sections {
		var sectionOptions []string
		for _, val := range section.SectionData {
			sectionOptions = append(sectionOptions, fmt.Sprintf(optionTemplate, val))
		}
		appendSection += fmt.Sprintf(sectionTemplate, section.SectionType, section.Name, strings.Join(sectionOptions, "\n"))
	}
	return fmt.Sprintf(haproxyTemplate, appendSection)
}

func (h Haproxy) CreateAPIServerConfig(localIp, port, balanceAlgorithm string, servers map[string]schema.NodeInformation) string {
	sectionOptions := []string{
		fmt.Sprintf("bind %s:%s", localIp, port),
		"mode tcp",
		"option tcplog",
		"option dontlognull",
		"option dontlog-normal",
		fmt.Sprintf("balance %s", balanceAlgorithm),
	}
	for _, node := range servers {
		sectionOptions = append(sectionOptions, fmt.Sprintf("server %s %s:6443 check inter 10s fall 2 rise 2 weight 1", node.SystemInfo.Node.Hostname, node.Ipv4DefaultIp))
	}
	return h.CreateConfig([]schema.ProxySection{
		{
			Name:        "kube-master",
			SectionType: "listen",
			SectionData: sectionOptions,
		},
	})
}

func (h Haproxy) GetSystemdServiceName() string {
	return "haproxy"
}
