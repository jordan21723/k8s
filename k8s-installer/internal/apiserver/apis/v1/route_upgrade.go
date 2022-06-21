package v1

import (
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/upgrade"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

var upgradeManage = []Route{
	{
		Path:           "/",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.ListUpgradableVersion,
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "List available upgrade version",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.UpgradableVersion{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.UpgradableVersion{},
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
		HTTPMethod:     http.MethodPost,
		Handler:        upgrade.CreateUpgradableVersion,
		ChallengeCode:  constants.ChallengeConfigUpgradeVersion,
		Doc:            "Create upgrade version",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.UpgradableVersion{},
		WriteDataModel: schema.UpgradableVersion{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
		Path:           "/{version-name}",
		HTTPMethod:     http.MethodPut,
		Handler:        upgrade.UpdateUpgradableVersion,
		ChallengeCode:  constants.ChallengeConfigUpgradeVersion,
		Doc:            "update upgrade version",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.UpgradableVersion{},
		WriteDataModel: schema.UpgradableVersion{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "version-name",
				Description:   "version name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
		Path:           "/{version-name}",
		HTTPMethod:     http.MethodDelete,
		Handler:        upgrade.DeleteUpgradableVersion,
		ChallengeCode:  constants.ChallengeConfigUpgradeVersion,
		Doc:            "Delete upgrade version",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "version-name",
				Description:   "version name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
		Path:           "/cluster-plan/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.ListUpgradePlane,
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "List specified cluster upgrade plan",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.UpgradePlan{},
		PathParams: []schema.Parameter{
			{
				Required:      false,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "cluster-id",
				Description:   "cluster id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradePlan{},
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
		Path:           "/cluster-plan/",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.ListUpgradePlane,
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "List all upgrade plan",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.UpgradePlan{},
		PathParams:     []schema.Parameter{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.UpgradePlan{},
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
		Path:           "/plan/{plan-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.UpgradePlanDetail,
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "show detail about an upgrade plan",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.UpgradePlan{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "plan-id",
				Description:   "plan id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
		Path:           "/plan/{plan-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        upgrade.DeleteUpgradePlan,
		ChallengeCode:  constants.ChallengeConfigUpgradeVersion,
		Doc:            "Delete an upgrade plan",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "plan-id",
				Description:   "plan id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
			{
				http.StatusPreconditionFailed,
				"Precondition check failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/plan/apply/{plan-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        upgrade.ApplyUpgradePlan,
		ChallengeCode:  constants.ChallengeUpgradeCluster,
		Doc:            "apply the plan do upgrade to cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "plan-id",
				Description:   "plan id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradableVersion{},
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
			{
				http.StatusPreconditionFailed,
				"Precondition check failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/make/{cluster-id}/{version-name}",
		HTTPMethod:     http.MethodPut,
		Handler:        upgrade.MakeUpgradePlan,
		ChallengeCode:  constants.ChallengeUpgradeCluster,
		Doc:            "Make a upgrade plan for specified cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.UpgradePlan{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "version-name",
				Description:   "version name",
				DataType:      "string",
				AllowMultiple: false,
			},
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "cluster-id",
				Description:   "cluster id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.UpgradePlan{},
			},
			{
				http.StatusBadRequest,
				"bad request",
				schema.HttpErrorResult{},
			},
			{
				http.StatusPreconditionFailed,
				"Precondition check failed",
				schema.HttpErrorResult{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
}
