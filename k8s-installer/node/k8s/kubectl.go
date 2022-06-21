package k8s

import (
	"bytes"
	"strings"
	"time"

	"k8s-installer/pkg/command"
	"k8s-installer/pkg/config/client"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/schema"
)

func ApplyCNIConfig(kubeadm schema.TaskKubeadm, config client.Config) error {
	saveTo := config.YamlDataDir + "/cni-config.yaml"
	log.Debugf("Try to save cni-config to location %s", saveTo)
	if err := util.WriteTxtToFile(saveTo, string(kubeadm.CNITemplate)); err != nil {
		log.Debugf("Failed to save kubeadmConfig to %s due to error %s", saveTo, err.Error())
		return err
	}

	log.Debugf("Apply cni config from %s", saveTo)
	if _, err := runKubectl([]string{"apply", "-f", saveTo}); err != nil {
		return nil
	}
	return nil
}

func KubectlExecutor(taskKubectl schema.TaskKubectl, config client.Config) (map[string]string, error) {
	if taskKubectl.Action == constants.ActionDelete {
		return nil, nil
	} else {
		if taskKubectl.WaitBeforeExecutor > 0 {
			go func() {
				log.Debugf("Wait %d before executor command", taskKubectl.WaitBeforeExecutor)
				time.Sleep(time.Duration(taskKubectl.WaitBeforeExecutor) * time.Second)
				kubectlCommand(taskKubectl, config)
			}()
			return nil, nil
		} else {
			return kubectlCommand(taskKubectl, config)
		}
	}
}

func kubectlCommand(taskKubectl schema.TaskKubectl, config client.Config) (map[string]string, error) {
	if taskKubectl.SubCommand == constants.KubectlSubCommandCreateOrApply {
		return doYaml(taskKubectl, config)
	} else if taskKubectl.SubCommand == constants.KubectlSubCommandDelete && len(taskKubectl.YamlTemplate) > 0 {
		return doYaml(taskKubectl, config)
	} else {
		return noneApplyOrCreate(taskKubectl)
	}
}

func noneApplyOrCreate(taskKubectl schema.TaskKubectl) (map[string]string, error) {
	results := map[string]string{}
	for _, args := range taskKubectl.CommandToRun {
		command := []string{taskKubectl.SubCommand}
		command = append(command, strings.Split(args, " ")...)
		if result, err := runKubectl(command); err != nil {
			log.Errorf("kubectl resulting in error state due to %s", err)
			if taskKubectl.CanIgnoreError {
				log.Debug("Task says this error can be ignore. So we move on...")
				continue
			}
			return nil, err
		} else {
			if taskKubectl.RequestResponse {
				if len(results) == 0 {
					results[constants.ReturnDataKeyKubectl] = result[constants.ReturnDataKeyKubectl]
				} else {
					results[constants.ReturnDataKeyKubectl] = results[constants.ReturnDataKeyKubectl] + " || " + result[constants.ReturnDataKeyKubectl]
				}
				return results, nil
			}
		}
	}
	return nil, nil
}

// apply/delete -f [yaml]
func doYaml(taskKubectl schema.TaskKubectl, config client.Config) (map[string]string, error) {
	results := map[string]string{}
	for fileName, yamlData := range taskKubectl.YamlTemplate {
		saveTo := config.YamlDataDir + "/" + fileName + ".yaml"
		log.Debugf("Try to save %s to location %s", fileName+".yaml", saveTo)
		if err := util.WriteTxtToFile(saveTo, string(yamlData)); err != nil {
			log.Debugf("Failed to save yaml template to %s due to error %s", saveTo, err.Error())
			return nil, err
		}
		if result, err := runKubectl([]string{taskKubectl.SubCommand, "-f", saveTo}); err != nil {
			log.Errorf("kubectl resulting in error state due to %s", err)
			if taskKubectl.CanIgnoreError {
				log.Debug("Task says this error can be ignore. So we move on...")
				continue
			}
			return nil, err
		} else {
			if taskKubectl.RequestResponse {
				if len(results) == 0 {
					results[constants.ReturnDataKeyKubectl] = result[constants.ReturnDataKeyKubectl]
				} else {
					results[constants.ReturnDataKeyKubectl] = results[constants.ReturnDataKeyKubectl] + " || " + result[constants.ReturnDataKeyKubectl]
				}
				return results, nil
			}
		}
	}
	return nil, nil
}

func runKubectl(args []string) (map[string]string, error) {
	result := map[string]string{}
	var stdErr, stdOut bytes.Buffer
	var err error
	log.Debugf("Run kubectl %s", strings.Join(args, " "))
	stdOut, stdErr, err = command.RunCmd("kubectl", args...)
	if err != nil {
		log.Errorf("Failed to run command %s due to following error:", "kubectl "+strings.Join(args, " "))
		log.Errorf("StdErr %s", stdErr.String())
		return nil, err
	} else {
		result[constants.ReturnDataKeyKubectl] = stdOut.String()
		log.Debugf("kubectl outputs: %s", stdOut.String())
	}
	return result, nil
}
