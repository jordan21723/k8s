package control_manager

import (
	"errors"
	"fmt"
	"k8s-installer/pkg/task_breaker"
	"net"
	"strconv"
	"strings"
	"time"

	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/network"
	"k8s-installer/schema"
)

func checkIpIsReachable(ip string, ipFamily string, timeoutSec time.Duration) error {
	runtimeCache := cache.GetCurrentCache()
	// try to reach client port
	connection, err := net.DialTimeout(ipFamily, ip+":"+strconv.FormatInt(int64(runtimeCache.GetServerRuntimeConfig(cache.NodeId).SignalPort), 10), timeoutSec)
	if err == nil {
		connection.Close()
	}
	return err
}

func toIPV6(ips []string) ([]net.IP, error) {
	var results []net.IP
	for _, ip := range ips {
		if !isValidIPV6(ip) {
			return []net.IP{}, errors.New(fmt.Sprintf("\"%s\" is not a valid ipv6 address", ip))
		}
		results = append(results, net.ParseIP(ip).To16())
	}
	return results, nil
}

func getAddressFromCIDR(cidr string) (string, error) {
	v4, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	return v4.String(), nil
}

func isValidIPV4(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ip.To4() != nil
}

func isValidIPV6(cidr string) bool {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ip.To16() != nil
}

func IsVipProperSet(cluster schema.Cluster) bool {
	if cluster.ExternalLB.ClusterVipV4 == "" {
		return false
	}
	return true
}

func CheckNodeExitsAndReachable(nodes []schema.ClusterNode, cluster schema.Cluster) error {
	errList, _ := checkNodeExitsAndReachable(nodes, cluster, true, false)
	if len(errList) > 0 {
		return errors.New(strings.Join(errList, " , "))
	}
	return nil
}

func CheckNodeReachable(nodeId string) error {
	runtimeCache := cache.GetCurrentCache()
	node, err := runtimeCache.GetNodeInformation(nodeId)
	currentNodeConfig := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	if err != nil {
		return err
	}
	if node.ProxyIpv4CIDR != "" {
		if err := network.CheckIpIsReachable(node.ProxyIpv4CIDR, currentNodeConfig.SignalPort, constants.ProtocolTcp, 1*time.Second); err != nil {
			return err
		}
	} else {
		if err := network.CheckIpIsReachable(node.Ipv4DefaultIp, currentNodeConfig.SignalPort, constants.ProtocolTcp, 1*time.Second); err != nil {
			return err
		}
	}

	return nil
}

func checkBackendStorageExistsAndReachable(cluster schema.Cluster) (errorList []string, warningList []string) {
	isEFKEnable := false
	isPGOEnable := false
	isOpenStackEnable := false
	isNFSEnable := false

	log.Debug("checkBackendStorageExistsAndReachable START")

	log.Debugf("cluster.PosterOperator is %#v", cluster.PostgresOperator)
	if cluster.PostgresOperator != nil {
		if cluster.PostgresOperator.Enable {
			isPGOEnable = true
		}
	}

	log.Debugf("cluster.CloudProvider is %#v", cluster.CloudProvider)
	if cluster.CloudProvider != nil {
		if cluster.CloudProvider.OpenStack != nil {
			if cluster.CloudProvider.OpenStack.Enable {
				isOpenStackEnable = true
			}
		}
	}
	log.Debugf("cluster.EFK is %#v", cluster.EFK)
	if cluster.EFK != nil {
		if cluster.EFK.Enable {
			isEFKEnable = true
		}
	}
	log.Debugf("cluster.Storage.NFS is %#v", cluster.Storage.NFS)
	if cluster.Storage.NFS != nil {
		if cluster.Storage.NFS.Enable {
			isNFSEnable = true
		}
	}
	/*
	   readme 81798, Algorithm refer to Karnaugh true table
	*/
	if isOpenStackEnable || isNFSEnable || (!isPGOEnable && !isEFKEnable) {
		log.Debug("The request for creating/recreating cluster meet the condition.")
	} else {
		log.Error("Please enable a backend storage(ex: OpenStack or NFS) for PGO or EFK")
		errorList = append(errorList, fmt.Sprintf("Please enable a backend storage(ex: OpenStack or NFS) for PGO or EFK"))
	}
	log.Debug("checkBackendStorageExistsAndReachable STOP")
	return
}

func checkNodeExitsAndReachable(nodes []schema.ClusterNode, cluster schema.Cluster, checkNetworkIntersect bool, isMaster bool) (errorList []string, warningList []string) {
	runtimeCache := cache.GetCurrentCache()
	nodeCollection, err := runtimeCache.GetNodeInformationCollection()
	if err != nil {
		errorList = append(errorList, fmt.Sprintf("Internal server error %s", err.Error()))
	}
	currentNodeConfig := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	for _, node := range nodes {
		log.Debugf("Get node id: %s from NodeCollection", node.NodeId)
		if foundNode, found := nodeCollection[node.NodeId]; !found {
			errorList = append(errorList, fmt.Sprintf("Node id %s not found in node collection", node.NodeId))
		} else {
			// check node region id proper set and match cluster region
			if foundNode.Region == nil {
				errorList = append(errorList, fmt.Sprintf("Node with id: %s region is not proper set, abort...", foundNode.Id))
			} else {
				if foundNode.Region.ID != cluster.Region {
					errorList = append(errorList, fmt.Sprintf("Node with id: %s region id '%s' does not match cluster region id '%s' , abort...", foundNode.Id, foundNode.Region.ID, cluster.Region))
				}
			}
			if isMaster {
				log.Debug("Check node has more than 1 core to act as a master?")
				if foundNode.SystemInfo.CPU.Threads <= 1 {
					msg := fmt.Sprintf("Node with id: %s only has 1 core cannot deploy master on it. Abort...", node.NodeId)
					log.Error(msg)
					errorList = append(errorList, msg)
				} else {
					log.Debugf("Node with id: %s has sufficient cores for deploy master on it.", node.NodeId)
				}
			}
			log.Debugf("Check node id: %s is taken?", node.NodeId)
			if foundNode.Role > 0 {
				log.Errorf("Node with id: %s is taken by cluster %s", node.NodeId, foundNode.BelongsToCluster)
				errorList = append(errorList, fmt.Sprintf("Node with id: %s is taken by cluster %s", node.NodeId, foundNode.BelongsToCluster))
			}
			log.Debugf("Check node id: %s is supported?", foundNode.Id)
			if foundNode.Supported {
				log.Debugf("Node with id: %s os %s %s is support", node.NodeId, foundNode.SystemInfo.OS.Vendor, foundNode.SystemInfo.OS.Version)
			} else {
				log.Errorf("Node with id: %s os %s %s is not support", node.NodeId, foundNode.SystemInfo.OS.Vendor, foundNode.SystemInfo.OS.Version)
				errorList = append(errorList, fmt.Sprintf("Node with id: %s os %s %s is not support", node.NodeId, foundNode.SystemInfo.OS.Vendor, foundNode.SystemInfo.OS.Version))
			}
			if foundNode.Status == constants.StateReady {
				log.Debugf("Node with id: %s current status is %s ", node.NodeId, foundNode.Status)
			} else {
				log.Errorf("Node with id: %s current status is %s. The Status has to be ready in order to proceed ", node.NodeId, foundNode.Status)
				errorList = append(errorList, fmt.Sprintf("Node with id: %s current status is %s. The Status has to be ready in order to proceed ", node.NodeId, foundNode.Status))
			}
			if foundNode.IsDisabled == false {
				log.Debugf("Node with id: %s is enabled", node.NodeId)
			} else {
				log.Errorf("Node with id: %s is disabled. It has to be enabled in order to proceed ", node.NodeId)
				errorList = append(errorList, fmt.Sprintf("Node with id: %s current status is %s. The Status has to be ready in order to proceed ", node.NodeId, foundNode.Status))
			}
			nodeCheckIp := foundNode.Ipv4DefaultIp
			if foundNode.ProxyIpv4CIDR != "" {
				log.Debugf("Found node proxy ipv4 cidr is set using proxy ip")
				nodeCheckIp = foundNode.ProxyIpv4CIDR
			}
			log.Debugf("Check node id: %s with ip %s is reachable?", node.NodeId, nodeCheckIp)
			if err := network.CheckIpIsReachable(nodeCheckIp, currentNodeConfig.SignalPort, constants.ProtocolTcp, 1*time.Second); err != nil {
				errorList = append(errorList, fmt.Sprintf("Node id: %s with ip %s runtime reachable check failed", node.NodeId, foundNode.Ipv4DefaultIp))
				log.Debugf("Failed to reach node id: %s with ip %s:%d", node.NodeId, nodeCheckIp, currentNodeConfig.SignalPort)
			} else {
				log.Debugf("Reach node id: %s with ip %s:%d", node.NodeId, nodeCheckIp, currentNodeConfig.SignalPort)
			}
			if checkNetworkIntersect && !cluster.Mock {
				errorList = append(errorList, validateNetworkIntersect(foundNode.Ipv4DefaultIp, foundNode.NodeIPV4AddressList, cluster.ControlPlane.ServiceV4CIDR, cluster.CNI.PodV4CIDR)...)
				if cluster.CNI.Calico.EnableDualStack {
					errorList = append(errorList, validateNetworkIntersect(foundNode.Ipv4DefaultIp, foundNode.NodeIPV6AddressList, cluster.ControlPlane.ServiceV6CIDR, cluster.CNI.PodV6CIDR)...)
				}
			}
		}
	}
	return
}

func validateNetworkIntersect(nodeIp string, netCIDRList []string, serviceNetCIDR, podNetCIDR string) []string {
	var errorList []string
	for _, cidr := range netCIDRList {
		// service cidr vs host network cidr
		if result, err := network.IntersectEachOther(cidr, serviceNetCIDR); err != nil {
			errorList = append(errorList, fmt.Sprintf("Failed to check node %s with network intersect between host cidr %s and service cidr %s due to error %s", nodeIp, cidr, serviceNetCIDR, err.Error()))
		} else if result {
			errorList = append(errorList, fmt.Sprintf("Network intersect validation failed between host cidr %s and service cidr %s of node %s", cidr, serviceNetCIDR, nodeIp))
		}
		// pod cidr vs host network cidr
		if result, err := network.IntersectEachOther(cidr, podNetCIDR); err != nil {
			errorList = append(errorList, fmt.Sprintf("Failed to check node %s with network intersect between host cidr %s and service cidr %s due to error %s", nodeIp, cidr, podNetCIDR, err.Error()))
		} else if result {
			errorList = append(errorList, fmt.Sprintf("Network intersect validation failed between host cidr %s and service cidr %s of node %s", cidr, podNetCIDR, nodeIp))
		}
	}
	// pod cidr vs service cidr
	if result, err := network.IntersectEachOther(serviceNetCIDR, podNetCIDR); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check node %s with network intersect between service cidr %s and pod cidr %s due to error %s", nodeIp, serviceNetCIDR, podNetCIDR, err.Error()))
	} else if result {
		errorList = append(errorList, fmt.Sprintf("Network intersect validation failed between service cidr %s and pod cidr %s of node %s", serviceNetCIDR, podNetCIDR, nodeIp))
	}
	return errorList
}

/*
check layout
1. ingress and external lb cannot coexist
2. ingress and work node cannot coexist because one haproxy is installed on each work node to use 127.0.0.1 as a vip for api server
*/

func checkNodeLayout(cluster schema.Cluster) []string {
	var errList []string
	shareNodes := map[string]byte{}

	foundMatch := false

	if len(cluster.Masters) < 1 {
		log.Debugf("master nodes require at least 1 node but got %d", len(cluster.Masters))
		errList = append(errList, fmt.Sprintf("master nodes require at least 1 node but got %d", len(cluster.Masters)))
	}

	if len(cluster.Workers) < 1 {
		log.Debugf("worker nodes require at least 1 node but got %d", len(cluster.Workers))
		errList = append(errList, fmt.Sprintf("worker nodes require at least 1 node but got %d", len(cluster.Workers)))
	}

	// find out nodes that master and worker not coexist
	for _, master := range cluster.Masters {
		shareNodes[master.NodeId] = '1'
		for _, worker := range cluster.Workers {
			shareNodes[worker.NodeId] = '1'
			if master.NodeId == worker.NodeId {
				foundMatch = true
			}
		}
		if foundMatch {
			foundMatch = false
			continue
		}
	}

	if cluster.ClusterLB != nil {

		if len(cluster.ClusterLB.Nodes) <= 1 {
			log.Debugf("lb nodes require at least 2 nodes but got %d", len(cluster.ClusterLB.Nodes))
			errList = append(errList, fmt.Sprintf("lb nodes require at least 2 nodes but got %d", len(cluster.ClusterLB.Nodes)))
		}
	}

	if cluster.Ingress != nil {
		for _, ingress := range cluster.Ingress.NodeIds {

			if _, found := shareNodes[ingress.NodeId]; !found {
				errList = append(errList, fmt.Sprintf("Ingress node with id %s must exist in master or work node list ", ingress.NodeId))
			}
			// check external lb
			if cluster.ClusterLB != nil {
				for _, node := range cluster.ClusterLB.Nodes {
					if ingress.NodeId == node.NodeId {
						log.Debugf("Ingress node and external lb node with id %s cannot coexist", ingress.NodeId)
						errList = append(errList, fmt.Sprintf("Ingress node and external lb node with id %s cannot coexist", ingress.NodeId))
					}
				}
			}

			// check work nodes which do not share with master
			// meaning one haproxy is installed on each work node to use 127.0.0.1 as a vip for api server
			/*		if _,found := shareNodes[ingress.NodeId];found {
					log.Debugf("Ingress node and worker node with id %s cannot coexist", ingress.NodeId)
					errList = append(errList, fmt.Sprintf("Ingress node and worker node with id %s cannot coexist", ingress.NodeId))
				}*/

		}
	}

	return errList
}

func PreCheck(cluster schema.Cluster) (canProceed bool, errorList []string, warningList []string, err error) {
	// check backend storages
	backendStorageErrList, backendStorageWarningList := checkBackendStorageExistsAndReachable(cluster)
	errorList = append(errorList, backendStorageErrList...)
	warningList = append(warningList, backendStorageWarningList...)
	// check masters
	nodeMasterErrList, nodeMasterWarningList := checkNodeExitsAndReachable(cluster.Masters, cluster, true, true)
	errorList = append(errorList, nodeMasterErrList...)
	warningList = append(warningList, nodeMasterWarningList...)
	// check workers
	nodeWorkerErrList, nodeWorkerWarningList := checkNodeExitsAndReachable(cluster.Workers, cluster, true, false)
	errorList = append(errorList, nodeWorkerErrList...)
	warningList = append(warningList, nodeWorkerWarningList...)
	// check ingress
	if cluster.Ingress != nil {
		nodeIngressErrList, nodeIngressWarningList := checkNodeExitsAndReachable(cluster.Ingress.NodeIds, cluster, false, false)
		errorList = append(errorList, nodeIngressErrList...)
		warningList = append(warningList, nodeIngressWarningList...)
	}
	// check external lb
	nodeExternalLBErrList, nodeExternalLBWarningList := checkNodeExitsAndReachable(cluster.ExternalLB.NodeIds, cluster, false, false)
	errorList = append(errorList, nodeExternalLBErrList...)
	warningList = append(warningList, nodeExternalLBWarningList...)

	// check k8s node layout
	errorList = append(errorList, checkNodeLayout(cluster)...)

	// check ks setting
	errorList = append(errorList, KsTargetHostClusterCheck(cluster)...)

	if err, errMsg := task_breaker.CheckAddonsDependencies(cluster); err != nil {
		errorList = append(errorList, errMsg...)
	}

	if strings.Count(cluster.ClusterName, ".") > 2 {
		errorList = append(errorList, fmt.Sprintf("Domain level cannot be greater than 2 e.g. %s", "[top level domain].[sub domain]"))
	}

	canProceed = len(errorList) == 0

	if err := network.CheckIpIsReachable(cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort, constants.ProtocolTcp, 1*time.Second); err != nil {
		log.Debugf("Private registry with address %s:%d is not reachable", cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort)
		warningList = append(warningList, fmt.Sprintf("Private registry with address %s:%d is not reachable", cluster.ContainerRuntime.PrivateRegistryAddress, cluster.ContainerRuntime.PrivateRegistryPort))
	}
	// naked return

	//errorList = PreCheckFileResource(architecture, errorList)
	return
}

func KsTargetHostClusterCheck(cluster schema.Cluster) []string {
	if cluster.KsClusterConf == nil || !cluster.KsClusterConf.Enabled || cluster.KsClusterConf.MultiClusterConfig.ClusterRole == constants.ClusterRoleHost {
		return nil
	}

	// MemberOfCluster must be set to a host cluster id
	if cluster.KsClusterConf.MemberOfCluster == "" {
		return []string{"Member cluster require a host cluster which is not provided"}
	}

	runtimeCache := cache.GetCurrentCache()
	targetHostCluster, err := runtimeCache.GetCluster(cluster.KsClusterConf.MemberOfCluster)
	if err != nil {
		return []string{fmt.Sprintf("Failed to find host cluster with id %s due to error: %s", cluster.KsClusterConf.MemberOfCluster, err.Error())}
	}

	if targetHostCluster == nil {
		return []string{fmt.Sprintf("Failed to find host cluster with id %s", cluster.KsClusterConf.MemberOfCluster)}
	}

	if targetHostCluster.Status != constants.ClusterStatusRunning {
		return []string{fmt.Sprintf("Failed to find host cluster with id %s is not ready yet ensure host cluster is in %s stat", cluster.KsClusterConf.MemberOfCluster, constants.ClusterStatusRunning)}
	}

	if targetHostCluster.KsClusterConf == nil || targetHostCluster.KsClusterConf.MultiClusterConfig.ClusterRole != constants.ClusterRoleHost {
		return []string{fmt.Sprintf("Host cluster with id %s is not host cluster", cluster.KsClusterConf.MemberOfCluster)}
	}

	return nil
}

//func PreCheckFileResource(errorList []string) []string {
//	runtimeCache := cache.GetCurrentCache()
//	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
//	for _, dep := range schema.Deps {
//		for osk, osv := range dep {
//			for kvk, kvv := range osv {
//				// for _, archv := range kvv {
//				for _, f := range kvv["x86_64"] {
//					//TODO: Add aarch64 check
//					fp := path.Join(config.ApiServer.ResourceServerFilePath, kvk, osk, "7", "x86_64", "package", f)
//					fe, _ := filepath.Glob(fp)
//					if len(fe) == 0 {
//						errorList = append(errorList, fmt.Sprintf("file %v do not exsist!", fp))
//					}
//				}
//				// }
//			}
//		}
//	}
//	return errorList
//}
