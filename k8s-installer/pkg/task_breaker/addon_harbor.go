package task_breaker

import (
	"fmt"
	"k8s-installer/pkg/config/server"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/task_breaker/utils"
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
	"reflect"
	"strings"
)

type (
	AddOnHarbor struct {
		Name         string
		TaskTimeOut  int
		enable       bool
		FullNodeList map[string]schema.NodeInformation
	}
)

func (params AddOnHarbor) DeployOrRemoveAddOn(operation *schema.Operation, cluster schema.Cluster, config server.Config, action string) error {
	if cluster.Action == constants.ActionCreate {
		certPath := "/etc/docker/certs.d/%s/%s.cert"
		keyPath := "/etc/docker/certs.d/%s/%s.key"
		caPath := "/etc/docker/certs.d/%s/ca.crt"
		// if cri is set to containerd redirect output to /etc/ssl/certs/
		if cluster.ContainerRuntime.CRIType == constants.CRITypeContainerd {
			certPath = "/etc/ssl/certs/%s/%s.cert"
			keyPath = "/etc/ssl/certs/%s/%s.key"
			caPath = "/etc/ssl/certs/%s/ca.crt"
		}
		if len(cluster.Harbor.Host) > 0 || cluster.Harbor.EnableTls {
			var harborHostnameConfigSteps []schema.NodeStep
			var configTlsCertSteps []schema.NodeStep
			// docker tls dir name default to user harbor ip
			for nodeId := range params.FullNodeList {
				// set /etc/hosts for harbor registry host only needed when cluster.Harbor.Host is not empty
				if len(cluster.Harbor.Host) > 0 {
					// set docker tls dir name to cluster.Harbor.Host
					harborHostnameConfigSteps = append(harborHostnameConfigSteps, schema.NodeStep{
						Id:               utils.GenNodeStepID(),
						Name:             "SetHostForHarborHost",
						NodeID:           nodeId,
						ServerMSGTimeOut: 5,
						Tasks: map[int]schema.ITask{
							0: schema.TaskSetHosts{
								TaskType: constants.TaskTypeSetHost,
								Action:   cluster.Action,
								TimeOut:  60,
								Hosts: map[string][]string{
									cluster.Harbor.Ip: {cluster.Harbor.Host},
								},
							},
						},
					})

				}
				if cluster.Harbor.EnableTls {

					var copyTo map[string][]byte
					copyTo = map[string][]byte{
						fmt.Sprintf(certPath, cluster.Harbor.Ip, cluster.Harbor.Ip): []byte(cluster.Harbor.Cert),
						fmt.Sprintf(keyPath, cluster.Harbor.Ip, cluster.Harbor.Ip):  []byte(cluster.Harbor.Key),
						fmt.Sprintf(caPath, cluster.Harbor.Ip):                      []byte(cluster.Harbor.CA),
					}
					if len(cluster.Harbor.Host) > 0 {
						copyTo[fmt.Sprintf(certPath, cluster.Harbor.Host, cluster.Harbor.Host)] = []byte(cluster.Harbor.Cert)
						copyTo[fmt.Sprintf(keyPath, cluster.Harbor.Host, cluster.Harbor.Host)] = []byte(cluster.Harbor.Key)
						copyTo[fmt.Sprintf(caPath, cluster.Harbor.Host)] = []byte(cluster.Harbor.CA)
					}
					configTlsCertSteps = append(configTlsCertSteps, schema.NodeStep{
						Id:     utils.GenNodeStepID(),
						Name:   "TaskTypeCopyTextFile-" + nodeId,
						NodeID: nodeId,
						Tasks: map[int]schema.ITask{
							0: schema.TaskCopyTextBaseFile{
								TaskType:  constants.TaskTypeCopyTextFile,
								TimeOut:   3,
								TextFiles: copyTo,
							},
						},
					})
				}
			}
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:                       "stepSetHostForHarborHost-" + operation.Id,
				Name:                     "stepSetHostForHarborHost",
				NodeSteps:                harborHostnameConfigSteps,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}

			if len(configTlsCertSteps) > 0 {
				operation.Step[len(operation.Step)] = &schema.Step{
					Id:                       "stepSetHarborTls-" + operation.Id,
					Name:                     "stepSetHarborTls",
					NodeSteps:                configTlsCertSteps,
					OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
					OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
				}
			}
		}
		// generator harbor k8s secret for ip
		GenKubeSecret("stepCreateHarborIPSecret", "harbor-k8s-installer-ip", cluster, operation, false)

		if len(cluster.Harbor.Host) > 0 {
			// generator harbor k8s secret for host
			GenKubeSecret("stepCreateHarborHostSecret", "harbor-k8s-installer-host", cluster, operation, true)
		}
	} else {
		// during cluster destroying
		// if you addons is none container based app you should do something to remove it
	}
	return nil

}

func GenKubeSecret(stepName, secretName string, cluster schema.Cluster, operation *schema.Operation, isHost bool) {
	address := cluster.Harbor.Ip
	if isHost {
		address = cluster.Harbor.Host
	}
	cmd := fmt.Sprintf("kubectl create secret docker-registry %s --docker-server=%s --docker-username=%s --docker-password=%s",
		secretName,
		fmt.Sprintf("%s:%d", address, cluster.Harbor.Port),
		cluster.Harbor.UserName,
		cluster.Harbor.Password)
	createKubeSecretStep := CommonRunWithSingleNode(cluster.Masters[0].NodeId, stepName, cluster.Action, 10, strings.Split(cmd, " "), false)
	operation.Step[len(operation.Step)] = &schema.Step{
		Id:                       fmt.Sprintf("%s-%s", stepName, operation.Id),
		Name:                     stepName,
		NodeSteps:                []schema.NodeStep{createKubeSecretStep},
		OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
		OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
	}
}

func (params AddOnHarbor) DeployOrRemoveAddonsWithCluster(operation *schema.Operation, cluster schema.Cluster, config server.Config) error {
	return params.DeployOrRemoveAddOn(operation, cluster, config, cluster.Action)
}

func (params AddOnHarbor) LivenessProbe(operationId string, cluster schema.Cluster, config server.Config) (schema.Step, error) {
	return schema.Step{
		Name: "SkipHarborCheck",
	}, nil
}

func (params AddOnHarbor) GetAddOnName() string {
	return params.Name
}

func (params AddOnHarbor) SetDataWithPlugin(plugAble plugable.IPlugAble, sumLicenseLabel uint16, cluster schema.Cluster) IAddOns {
	if plugAble == nil || reflect.ValueOf(plugAble).IsNil() {
		params.enable = false
	} else {
		// do not look into license switch
		params.enable = plugAble.IsEnable()
	}
	return params
}

func (params AddOnHarbor) DeployOrRemoveAddonsWithNodeAddOrRemove(operation *schema.Operation, cluster schema.Cluster, config server.Config, nodeAction string, nodesToAddOrRemove []schema.ClusterNode) error {
	if nodeAction == constants.ActionCreate {
		certPath := "/etc/docker/certs.d/%s/%s.cert"
		keyPath := "/etc/docker/certs.d/%s/%s.key"
		caPath := "/etc/docker/certs.d/%s/ca.crt"
		// if cri is set to containerd redirect output to /etc/ssl/certs/
		if cluster.ContainerRuntime.CRIType == constants.CRITypeContainerd {
			certPath = "/etc/ssl/certs/%s/%s.cert"
			keyPath = "/etc/ssl/certs/%s/%s.key"
			caPath = "/etc/ssl/certs/%s/ca.crt"
		}
		if len(cluster.Harbor.Host) > 0 || cluster.Harbor.EnableTls {
			var harborHostnameConfigSteps []schema.NodeStep
			var configTlsCertSteps []schema.NodeStep
			// docker tls dir name default to user harbor ip
			for _, node := range nodesToAddOrRemove {
				// set /etc/hosts for harbor registry host only needed when cluster.Harbor.Host is not empty
				if len(cluster.Harbor.Host) > 0 {
					// set docker tls dir name to cluster.Harbor.Host
					harborHostnameConfigSteps = append(harborHostnameConfigSteps, schema.NodeStep{
						Id:               utils.GenNodeStepID(),
						Name:             "SetHostForHarborHost",
						NodeID:           node.NodeId,
						ServerMSGTimeOut: 5,
						Tasks: map[int]schema.ITask{
							0: schema.TaskSetHosts{
								TaskType: constants.TaskTypeSetHost,
								Action:   cluster.Action,
								TimeOut:  60,
								Hosts: map[string][]string{
									cluster.Harbor.Ip: {cluster.Harbor.Host},
								},
							},
						},
					})

				}
				if cluster.Harbor.EnableTls {
					var copyTo map[string][]byte
					copyTo = map[string][]byte{
						fmt.Sprintf(certPath, cluster.Harbor.Ip, cluster.Harbor.Ip): []byte(cluster.Harbor.Cert),
						fmt.Sprintf(keyPath, cluster.Harbor.Ip, cluster.Harbor.Ip):  []byte(cluster.Harbor.Key),
						fmt.Sprintf(caPath, cluster.Harbor.Ip):                      []byte(cluster.Harbor.CA),
					}
					if len(cluster.Harbor.Host) > 0 {
						copyTo[fmt.Sprintf(certPath, cluster.Harbor.Host, cluster.Harbor.Host)] = []byte(cluster.Harbor.Cert)
						copyTo[fmt.Sprintf(keyPath, cluster.Harbor.Host, cluster.Harbor.Host)] = []byte(cluster.Harbor.Key)
						copyTo[fmt.Sprintf(caPath, cluster.Harbor.Host)] = []byte(cluster.Harbor.CA)
					}
					configTlsCertSteps = append(configTlsCertSteps, schema.NodeStep{
						Id:     utils.GenNodeStepID(),
						Name:   "TaskTypeCopyTextFile-" + node.NodeId,
						NodeID: node.NodeId,
						Tasks: map[int]schema.ITask{
							0: schema.TaskCopyTextBaseFile{
								TaskType:  constants.TaskTypeCopyTextFile,
								TimeOut:   3,
								TextFiles: copyTo,
							},
						},
					})

				}
			}
			operation.Step[len(operation.Step)] = &schema.Step{
				Id:                       "stepSetHostForHarborHost-" + operation.Id,
				Name:                     "stepSetHostForHarborHost",
				NodeSteps:                harborHostnameConfigSteps,
				OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
				OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
			}

			if len(configTlsCertSteps) > 0 {
				operation.Step[len(operation.Step)] = &schema.Step{
					Id:                       "stepSetHarborTls-" + operation.Id,
					Name:                     "stepSetHarborTls",
					NodeSteps:                configTlsCertSteps,
					OnStepDoneOrErrorHandler: OnErrorIgnoreHandler{},
					OnStepTimeOutHandler:     OnTimeOutIgnoreHandler{},
				}
			}

		}
	}
	return nil
}

func (params AddOnHarbor) IsEnable() bool {
	return params.enable
}
