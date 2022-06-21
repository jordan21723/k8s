package block_device

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/disk"
	"k8s-installer/pkg/command"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
)

func IsMountPoint(file string) (bool, error) {
	stat, err := os.Stat(file)
	if err != nil {
		return false, err
	}
	rootStat, err := os.Lstat(file + "/..")
	if err != nil {
		return false, err
	}
	// If the directory has the same device as parent, then it's not a mountpoint.
	return stat.Sys().(*syscall.Stat_t).Dev != rootStat.Sys().(*syscall.Stat_t).Dev, nil
}

func Mount(dev string, targetDir string, fsType string, enableBootCheck bool, umount bool, force bool, backupFilePath string) error {
	if !umount {
		mountList, err := disk.Partitions(true)

		if err != nil {
			return err
		}

		for _, mnt := range mountList {
			if strings.Contains(mnt.Device, dev) {
				return errors.New(fmt.Sprintf("Device %s already mount to dir %s", dev, mnt.Mountpoint))
			}
		}
	}

	mntCommand := "mount"
	args := []string{}

	if umount {
		mntCommand = "umount"
		if force {
			args = append(args, "-f")
		}
		args = append(args, dev)
	} else {
		args = append(args, "-t", fsType, dev, targetDir)
	}

	var stdErr bytes.Buffer
	var errCMD error
	// install yum-utils
	log.Debugf("Attempting to %s device %s", mntCommand, dev)
	_, stdErr, errCMD = command.RunCmd(mntCommand, args...)
	if errCMD != nil {
		log.Errorf("Failed to mount device %s due to following error:", dev)
		log.Errorf("StdErr %v", stdErr)
		return errCMD
	}

	return WriteFstab(dev, targetDir, fsType, enableBootCheck, umount, backupFilePath)
}

func WriteFstab(dev string, targetDir string, fsType string, enableBootCheck bool, umount bool, backupFilePath string) error {

	newId := uuid.New().String()
	dest := backupFilePath + "/_backup_" + newId
	log.Debugf("Attempting to backup /etc/fstab to %s", dest)
	if err := fileutils.CopyFile("/etc/fstab", dest); err != nil {
		return err
	}

	fstabString := "%s %s %s defaults 0 %d\n"
	bootCheck := 0
	if enableBootCheck {
		bootCheck = 1
	}
	if umount {
		regxExpression := fmt.Sprintf(`^%s`, dev)
		rex := regexp.MustCompile(regxExpression)
		dat, err := ioutil.ReadFile("/etc/fstab")
		if err != nil {
			return err
		}
		arr := strings.Split(string(dat), "\n")
		result := ""
		for _, line := range arr {
			if rex.MatchString(line) {
				continue
			}
			result += line + "\n"
		}
		if err := util.WriteTxtToFile("/etc/fstab", result); err != nil {
			return err
		}
	} else {
		log.Debug("Attempting to edit /etc/fstab")
		fstabString = fmt.Sprintf(fstabString, dev, targetDir, fsType, bootCheck)
		if err := util.AppendTxtToFile("/etc/fstab", fstabString, 0644); err != nil {
			return err
		}
	}
	return nil
}
