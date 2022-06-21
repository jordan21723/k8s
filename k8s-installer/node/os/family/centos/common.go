package centos

import (
	"bytes"
	"strings"

	"k8s-installer/pkg/command"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
)

func DisableSelinux() error {
	var stdErr bytes.Buffer
	var err error

	// run setenforce 0
	log.Debug("Try to run command setenforce 0")
	_, stdErr, err = command.RunCmd("setenforce", "0")

	if err != nil {
		log.Error("Failed to run command setenforce 0 due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	// set /etc/selinux/config to disable
	log.Debug("Try to run command sed -i s/^SELINUX=.*$/SELINUX=disabled/ /etc/selinux/config")
	_, stdErr, err = command.RunCmd("sed", "-i", "s/^SELINUX=.*$/SELINUX=disabled/", "/etc/selinux/config")
	if err != nil {
		log.Error("Failed to disable selinux due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	log.Debug("Successfully change selinux config to disable")
	return nil
}

func UninstallPackage(packages []string) error {
	var stdErr bytes.Buffer
	var err error
	subCommand := []string{"-e", "--nodeps"}
	subCommand = append(subCommand, packages...)
	// run setenforce 0
	log.Debug("Try to run command yum remove")
	_, stdErr, err = command.RunCmd("rpm", subCommand...)
	if err != nil {
		log.Warnf("Failed to removed all package due to following error:")
		log.Warnf("StdErr %s", stdErr.String())
	}
	log.Debugf("Successfully removed all package %s", strings.Join(packages, ","))
	return err
}

func OnlineInstallK8sYumDep(enableIpvs bool) error {
	kubernetesRepo := `[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl`
	var stdErr bytes.Buffer
	var err error
	log.Debug("Try to set repo...")
	if err := util.WriteTxtToFile("/etc/yum.repos.d/kubernetes.repo", kubernetesRepo); err != nil {
		return err
	}

	if enableIpvs {
		log.Debug("Try to run command yum install -y kubelet kubeadm kubectl ebtables ipvsadm --disableexcludes=kubernetes")
		_, stdErr, err = command.RunCmd("yum", "install", "-y", "kubelet", "kubeadm", "kubectl", "ebtables", "ipvsadm", "--disableexcludes=kubernetes")
	} else {
		log.Debug("Try to run command yum install -y kubelet kubeadm kubectl ebtables --disableexcludes=kubernetes")
		_, stdErr, err = command.RunCmd("yum", "install", "-y", "kubelet", "kubeadm", "kubectl", "ebtables", "--disableexcludes=kubernetes")
	}

	if err != nil {
		log.Error("Failed to run command yum install -y kubelet kubeadm kubectl ebtables --disableexcludes=kubernetes due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}
	return err
}
