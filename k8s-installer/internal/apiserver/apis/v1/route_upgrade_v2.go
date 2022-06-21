package v1

import (
	"net/http"

	"k8s-installer/internal/apiserver/controller/upgrade"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var upgradeV2Manage = []Route{
	{
		Path:           "/upgrades/version",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.ListUpgradableVersion,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades/version",
		HTTPMethod:     http.MethodPost,
		Handler:        upgrade.CreateUpgradableVersion,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades//version/{version-name}",
		HTTPMethod:     http.MethodPut,
		Handler:        upgrade.UpdateUpgradableVersion,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades//version/{version-name}",
		HTTPMethod:     http.MethodDelete,
		Handler:        upgrade.DeleteUpgradableVersion,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades/plans",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.ListUpgradePlaneV2,
		Tags:           []string{"Core_Upgrade"},
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "List specified cluster upgrade plan",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.UpgradePlan{},
		QueryParams: []schema.Parameter{
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
		Path:           "/upgrades/plans/{plan-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        upgrade.UpgradePlanDetail,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades/plans/{plan-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        upgrade.DeleteUpgradePlan,
		Tags:           []string{"Core_Upgrade"},
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
	// apply plan
	{
		Path:           "/upgrades/plans/{plan-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        upgrade.ApplyUpgradePlan,
		Tags:           []string{"Core_Upgrade"},
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
		Path:           "/upgrades/plans",
		HTTPMethod:     http.MethodPost,
		Handler:        upgrade.MakeUpgradePlanV2,
		Tags:           []string{"Core_Upgrade"},
		ChallengeCode:  constants.ChallengeUpgradeCluster,
		Doc:            "Make a upgrade plan for specified cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.UpgradePlan{},
		QueryParams: []schema.Parameter{
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
