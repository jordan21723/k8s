package license

import (
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/license"
	"net/http"

	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"

	"github.com/emicklei/go-restful"
)

func CreateLicense(request *restful.Request, response *restful.Response) {

	lic := schema.LicenseInfo{}
	if err := request.ReadEntity(&lic); err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}

	systemInfo, err := fileutils.ParseLicense(lic.License)
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   err.Error(),
		})
		return
	}

	if !systemInfo.LicenseValid {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: constants.ErrLicenseInvalid,
			Message:   "license invalid",
		})
		return
	}

	runtimeCache := cache.GetCurrentCache()
	err = runtimeCache.CreateOrUpdadeLicense(lic)
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   err.Error(),
		})
		return
	}

	systemInfo.License = lic.License
	response.WriteAsJson(&systemInfo)
}

func GetLicense(request *restful.Request, response *restful.Response) {

	runtimeCache := cache.GetCurrentCache()
	license, err := runtimeCache.GetLicense()
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   err.Error(),
		})
		return
	}

	if license == nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusNotFound,
			Message:   "License not found!",
		})
		return
	}

	systemInfo, err := fileutils.ParseLicense(license.License)
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   err.Error(),
		})
		return
	}
	systemInfo.License = license.License

	response.WriteAsJson(&systemInfo)
}

func GetAddonsFromLicense(request *restful.Request, response *restful.Response) {
	if sumLicenseLabel, err := license.SumLicenseLabel(); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	} else if constants.LLConsole&sumLicenseLabel == constants.LLConsole {
		response.WriteAsJson(map[string]byte{"middle-platform": 0, "console": 0, "pgo": 0})
	} else {
		response.WriteAsJson(map[string]byte{})
	}
}
