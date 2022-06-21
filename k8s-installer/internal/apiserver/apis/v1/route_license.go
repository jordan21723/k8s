package v1

import (
	"k8s-installer/pkg/constants"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/license"
	"k8s-installer/schema"
)

var licenseManage = []Route{
	{
		Path:           "/",
		HTTPMethod:     http.MethodPost,
		Handler:        license.CreateLicense,
		ChallengeCode:  constants.ChallengeCodeManageCert,
		Doc:            "Create or Update License",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.LicenseInfo{},
		WriteDataModel: schema.SystemInfo{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.SystemInfo{},
			},
			{
				http.StatusBadRequest,
				"bad request",
				schema.HttpErrorResult{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/",
		HTTPMethod:     http.MethodGet,
		Handler:        license.GetLicense,
		ChallengeCode:  constants.ChallengeCodeManageCert,
		Doc:            "Get License",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.SystemInfo{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.SystemInfo{},
			},
			{
				http.StatusBadRequest,
				"bad request",
				schema.HttpErrorResult{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:          "/addons",
		HTTPMethod:    http.MethodGet,
		Handler:       license.GetAddonsFromLicense,
		ChallengeCode: 0,
		Doc:           "Get addons from license",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: nil,
		WriteDataModel: struct {
			Console        int `json:"console" description:"Is console enabled 0=disable 1=enabled"`
			MiddlePlatform int `json:"middle_platform" description:"Is middle platform enabled 0=disable 1=enabled"`
			Pgo            int `json:"pgo" description:"Is middle platform enabled 0=disable 1=enabled"`
		}{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
				ReturnModel: struct {
					Console        int `json:"console" description:"Is console enabled 0=disable 1=enabled"`
					MiddlePlatform int `json:"middle_platform" description:"Is middle platform enabled 0=disable 1=enabled"`
					Pgo            int `json:"pgo" description:"Is middle platform enabled 0=disable 1=enabled"`
				}{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
}
