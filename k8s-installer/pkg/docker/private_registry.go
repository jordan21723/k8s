package docker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"k8s-installer/pkg/util"
)

type PRV2Repositories struct {
	Repositories []string `json:"repositories"`
}

type PRV2Image struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type PRV2Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func ListImagesV2(registryAddress string) (map[string]PRV2Image, error) {
	getRegistryUrl := registryAddress + "/v2/_catalog?n=2000"
	respBody, status, err := util.CommonRequest(getRegistryUrl, http.MethodGet, "", json.RawMessage{}, nil, true, true, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("Unable to list image from %s due to error: %s. Is that a pubilic registry? Is network reachable?", registryAddress, err.Error())
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("Unable to list image from %s due to unknow reason", registryAddress)
	}
	var PRV2Repositories PRV2Repositories
	if err := json.Unmarshal(respBody, &PRV2Repositories); err != nil {
		return nil, err
	}
	var PRV2Images = map[string]PRV2Image{}
	for _, repo := range PRV2Repositories.Repositories {
		getTagUrl := registryAddress + "/v2/" + repo + "/tags/list"
		var PRV2Image PRV2Image
		if respBody, status, err := util.CommonRequest(getTagUrl, http.MethodGet, "", json.RawMessage{}, nil, true, true, 2*time.Second); err != nil {
			return nil, err
		} else if status != http.StatusOK {
			return nil, fmt.Errorf("Unable to list image tags from %s due to unknow reason", registryAddress)
		} else {
			if err := json.Unmarshal(respBody, &PRV2Image); err != nil {
				return nil, err
			}
			PRV2Images[PRV2Image.Name] = PRV2Image
		}
	}
	return PRV2Images, nil
}
