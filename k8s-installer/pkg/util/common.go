package util

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"k8s-installer/pkg/command"
	"k8s-installer/pkg/log"

	"github.com/spf13/cast"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func StringToInt(value string) (int, error) {
	if val, err := strconv.Atoi(value); err != nil {
		return -1, err
	} else {
		return val, nil
	}
}

func IntGreaterThan(a, b int) bool {
	return a > b
}

func IntGreaterEqualThan(a, b int) bool {
	return a >= b
}

func IntLessThan(a, b int) bool {
	return a < b
}

func IntLessEqualThan(a, b int) bool {
	return a <= b
}

func StringSliceToMap(list []string) map[string]byte {
	result := map[string]byte{}
	for _, val := range list {
		result[val] = '0'
	}
	return result
}

/**
 * If first string is empty, return second one.
 */
func StringDefaultIfNotSet(str, defaultStr string) string {
	if str != "" {
		return str
	}
	return defaultStr
}

/**
 * If first value is 0, return second one.
 */
func IntPDefaultIfNotSet(val *int, defaultVal int) *int {
	if val != nil {
		return val
	}
	return &defaultVal
}

func IntDefaultIfZero(val int, defaultVal int) int {
	if val == 0 {
		return defaultVal
	}
	return val
}

/**
 * Return the BASE64 encoding of the given string.
 */
func Base64Encode(pre string) string {
	return base64.StdEncoding.EncodeToString([]byte(pre))
}

/**
 * Split slices an input string into all substrings separated by delimiter.
 */
func StringSplit(sep string, in interface{}) ([]string, error) {
	str, err := cast.ToStringE(in)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(str, sep), nil
}

/**
 * Locate elements in string array by index.
 */
func StringArrayLocate(index int, arr interface{}) (string, error) {
	set, err := cast.ToStringSliceE(arr)
	if err != nil {
		return "", err
	}
	if index < 0 || int(index) > len(set)-1 {
		return "", fmt.Errorf("index %d is out of array range", index)
	}
	return set[index], nil
}

// Name: cri-tools
// BuildRequires: systemd
// BuildRequires: curl

// Name: kubeadm
// BuildRequires: systemd
// BuildRequires: curl
// Requires: kubelet >= {{ index .Dependencies "kubelet" }}
// Requires: kubectl >= {{ index .Dependencies "kubectl" }}
// Requires: kubernetes-cni >= {{ index .Dependencies "kubernetes-cni" }}
// Requires: cri-tools >= {{ index .Dependencies "cri-tools" }}

// Name: kubectl
// BuildRequires: systemd
// BuildRequires: curl

// Name: kubelet
// BuildRequires: systemd
// BuildRequires: curl
// Requires: iptables >= 1.4.21
// Requires: kubernetes-cni >= {{ index .Dependencies "kubernetes-cni" }}
// Requires: socat
// Requires: util-linux
// Requires: ethtool
// Requires: iproute
// Requires: ebtables
// Requires: conntrack

// Name: kubernetes-cni
// BuildRequires: systemd
// BuildRequires: curl
// Requires: kubelet

func RemoveInstalledRPM() error {
	out, stdErr, err := command.RunCmd("rpm", "-qa")
	if err != nil {
		log.Errorf("Failed to get all kubernetes package due to following error:")
		log.Errorf("StdErr %s", stdErr.String())
		return err
	}

	rpmList := strings.Split(out.String(), "\n")

	k8sRpmList := []string{"cri-tools", "kubeadm", "kubectl", "kubelet", "kubernetes-cni"}

	for _, rpm := range rpmList {
		for _, k8sRpm := range k8sRpmList {
			if strings.Contains(rpm, k8sRpm) {
				log.Debug("UnInstalling rpm", rpm)
				_, stdErr, err = command.RunCmd("bash", "-c", fmt.Sprintf("rpm -e --nodeps %s", rpm))
				if err != nil {
					log.Errorf("Failed to uninstall %s due to following error:", rpm)
					log.Errorf("StdErr %s", stdErr.String())
					return err
				}
			}
		}
	}

	return nil
}

func GenerateLBKubeConfig(vip string, clusterId string) (*clientcmdapi.Config, error) {
	log.Debug("Load kubeconfig from", fmt.Sprintf("/root/.kube/config%v", clusterId))
	currentConfig, err := clientcmd.LoadFromFile(fmt.Sprintf("/root/.kube/config%v", clusterId))
	if err != nil {
		log.Error("Load kube config error", err)
		return nil, err
	}
	currentCtx, exists := currentConfig.Contexts[currentConfig.CurrentContext]
	if !exists {
		log.Debugf("failed to get current context")
		return nil, fmt.Errorf("failed to find CurrentContext in Contexts of the kubeconfig file")
	}

	log.Debug("Get config cluster", currentConfig.Clusters)
	_, exists = currentConfig.Clusters[currentCtx.Cluster]
	if exists {
		currentConfig.Clusters[currentCtx.Cluster].Server = fmt.Sprintf("https://%v:6443", vip)
	}
	return currentConfig, nil
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
