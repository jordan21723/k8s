package util

import (
	"strconv"
	"strings"
)

// only compare 3 digital 1.18.10 1.19.2
func NewerK8sVersion(version1, version2 string) (string, error) {
	versionDec1 := strings.Split(version1, ".")
	versionDec2 := strings.Split(version2, ".")
	for i := 0; i < 3; i++ {
		v1Dec, errV1 := strconv.ParseInt(versionDec1[i], 10, 64)
		if errV1 != nil {
			return "", errV1
		}
		v2Dec, errV2 := strconv.ParseInt(versionDec2[i], 10, 64)
		if errV2 != nil {
			return "", errV2
		}
		if v1Dec == v2Dec {
			continue
		}
		if v1Dec > v2Dec {
			return version1, nil
		} else {
			return version2, nil
		}
	}
	return "", nil
}
