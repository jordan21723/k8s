package license

import (
	"errors"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"
	"strings"
)

func getOrLoadLicenseInfo() (*schema.SystemInfo, error) {
	// always load from db in that case license update does not require server restart
	runtimeCache := cache.GetCurrentCache()
	license, err := runtimeCache.GetLicense()
	if err != nil {
		log.Errorf("Failed to pre-loading license info due to error: %v", err)
		return nil, err
	}
	if license == nil {
		return nil, errors.New("Unable to pre-loading license info due to license not found in db ")
	}
	systemInfo, err := fileutils.ParseLicense(license.License)
	if err != nil {
		log.Errorf("Failed to parse license due to error: %v", err)
		return nil, err
	}

	return systemInfo, nil
}

func SumLicenseLabel() (uint16, error) {
	license, err := getOrLoadLicenseInfo()
	if err != nil {
		return 0, err
	}
	// debug modules is empty scenario
	//license.Modules = ""
	addonsList := strings.Split(license.Modules, ",")
	var result uint16 = 0
	for _, name := range addonsList {
		if _, exist := constants.LicenseLabelList[name]; exist {
			result |= constants.LicenseLabelList[name]
		}
	}
	return result, nil
}
