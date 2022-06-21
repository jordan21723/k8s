package block_device

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/log"
	"strings"
)

// check whether the device is mounted already
func DeviceIsReady(dev string) (bool, error) {
	mountList, err := disk.Partitions(true)

	if err != nil {
		return false, err
	}

	for _, mnt := range mountList {
		if strings.Contains(mnt.Device, dev) {
			return false, nil
		}
	}

	return true, nil
}

func CheckDirWithMountSize(dir string, size uint) (bool, error) {
	result, err := IsMountPoint(dir)
	if err != nil {
		return false, err
	}

	if !result {
		return false, nil
	}

	partition, err := disk.Usage(dir)

	if err != nil {
		return false, err
	} else if partition.Free >= uint64(size*1024*1024*1024) {
		return true, nil
	} else {
		return false, nil
	}
}

func MakeFS(dev string, force bool, fsType string) error {
	tools := "mkfs.%s"
	switch fsType {
	case BTRFS:
	case EXT2:
	case EXT3:
	case EXT4:
	case XFS:
		tools = fmt.Sprintf(tools, fsType)
	default:
		return errors.New(fmt.Sprintf("File system format %s is not a support format", fsType))
	}

	var _, stdErr bytes.Buffer
	var err error
	log.Debug("Attempt to make fs")
	if force {
		_, stdErr, err = command.RunCmd(tools, "-f", dev)
	} else {
		_, stdErr, err = command.RunCmd(tools, dev)
	}
	if err != nil {
		log.Errorf("Failed to mkfs.%s %s due to error %s", fsType, dev, stdErr.String())
		return err
	}

	return nil
}
