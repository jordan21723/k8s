package container_runtime

import (
	"fmt"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util/fileutils"
	"path"
)

func LoadContainerRuntimeTar(localImagePath string, containerRuntime string) error {
	if localImagePath != "" {
		log.Debugf("Detect localImagePath is set %s go with edge mode by loading container runtime image locally.", localImagePath)
		fileList, err := fileutils.WalkMatch(localImagePath, "*.tar")
		if err != nil {
			log.Errorf("Failed to load file in dir %s due to error %s", localImagePath, err.Error())
			return err
		}
		if len(fileList) == 0 {
			log.Errorf("Dir %s contain no file at all, skip pre load image", localImagePath)
			return nil
		}

		var commands string
		switch containerRuntime {
		case constants.CRITypeContainerd:
			commands = `ctr -n=k8s.io --address=/run/containerd/containerd.sock image import `
		case constants.CRITypeDocker:
			commands = `docker image load --input `
		default:
			err = fmt.Errorf("Container runtime %s currently not support yet", containerRuntime)
			return err
		}

		for _, f := range fileList {
			loadCommands := commands + path.Join(localImagePath, f)
			log.Debugf("Load %s from local", f)
			_, stdErr, err := command.RunCmd("bash", "-c", loadCommands)
			if err != nil {
				log.Errorf("Failed to load local image due to following error:")
				log.Errorf("StdErr %s", stdErr.String())
				return err
			}
		}
	} else {
		log.Debugf("LocalImagePath is set empty. We consider user doesn't require pre load image on purpose, let's skip pre load image then")
	}
	return nil
}
