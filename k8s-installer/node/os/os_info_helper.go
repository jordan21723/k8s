package os

import (
	"os/user"
	"strconv"
	"strings"

	"github.com/zcalusic/sysinfo"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"
)

func GetAllSystemInformation() (*sysinfo.SysInfo, error) {
	current, err := user.Current()
	if err != nil {
		return nil, err
	}
	if current.Uid != "0" {
		log.Fatal("requires superuser privilege")
	}
	sysInfo := &sysinfo.SysInfo{}
	sysInfo.GetSysInfo()
	return sysInfo, nil
}

func GetOSFamily() (string, error) {
	sysInfo := &sysinfo.SysInfo{}
	var err error
	if sysInfo, err = GetAllSystemInformation(); err != nil {
		return "", err
	}
	return sysInfo.OS.Vendor, nil
}

func ValidationOsFamily(families []string) (bool, error) {
	sysInfo := &sysinfo.SysInfo{}
	var err error
	if sysInfo, err = GetAllSystemInformation(); err != nil {
		return false, err
	}
	for _, supportedOS := range families {
		if sysInfo.OS.Vendor == supportedOS {
			return true, nil
		}
	}
	return false, nil
}

func ValidationOsFamilyWithGivenSysInfo(families, versions []string, givenSysInfo sysinfo.SysInfo, checkVersion bool) bool {
	for _, supportedOS := range families {
		if givenSysInfo.OS.Vendor == supportedOS {
			if checkVersion {
				for _, version := range versions {
					if version == givenSysInfo.OS.Version {
						return true
					}
				}
			} else {
				return true
			}
		}
	}
	return false
}

func KernelVersionMustGreaterEqualThan(major, sub, version int) (bool, error) {
	osMajor, osSub, osVer, err := GetKernelVersion()
	if err != nil {
		return false, err
	}
	if util.IntGreaterThan(osMajor, major) {
		return true, nil
	} else if major == osMajor {
		if util.IntGreaterThan(osSub, sub) {
			return true, nil
		} else if sub == osSub {
			if util.IntGreaterEqualThan(osVer, version) {
				return true, nil
			} else {
				return false, nil
			}
		} else {
			return false, nil
		}
	} else {
		return false, nil
	}
}

func GetKernelVersion() (major, sub, version int, err error) {
	sysInfo := &sysinfo.SysInfo{}
	if sysInfo, err = GetAllSystemInformation(); err != nil {
		return 0, 0, 0, err
	}
	versionString := strings.Split(sysInfo.Kernel.Release, "-")[0]
	versionSlice := strings.Split(versionString, ".")
	if val, err := strconv.Atoi(versionSlice[0]); err != nil {
		log.Fatal(err)
	} else {
		major = val
	}

	if val, err := strconv.Atoi(versionSlice[1]); err != nil {
		log.Fatal(err)
	} else {
		sub = val
	}

	if val, err := strconv.Atoi(versionSlice[2]); err != nil {
		log.Fatal(err)
	} else {
		version = val
	}
	return
}
