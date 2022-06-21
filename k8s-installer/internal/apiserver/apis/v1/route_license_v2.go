package v1

import (
	"k8s-installer/internal/apiserver/controller/license"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var licenseV2Manage = []Route{
	{
		Path:           "/licenses",
		HTTPMethod:     http.MethodPost,
		Handler:        license.CreateLicense,
		Tags:           []string{"Core_License"},
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
		Path:           "/licenses",
		HTTPMethod:     http.MethodGet,
		Handler:        license.GetLicense,
		Tags:           []string{"Core_License"},
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
		Path:           "/licenses/addons",
		HTTPMethod:     http.MethodGet,
		Tags:           []string{"Core_License"},
		Handler:        license.GetAddonsFromLicense,
		ChallengeCode:  0,
		Doc:            "Get addons from license",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: map[string]byte{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.SystemInfo{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
}
