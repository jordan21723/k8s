package v1

import (
	"k8s-installer/internal/apiserver/controller/operation"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var operationV2Manager = []Route{
	{
		Path:           "/operation/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        operation.ListOperation,
		Tags:           []string{"Core_Operation"},
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "List cluster operation",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Operation{},
		PathParams: []schema.Parameter{
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
				ReturnModel: []schema.Operation{},
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
		Path:           "/operation/upgrade/{upgrade-plan-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        operation.ListOperationWithUpgradePlan,
		Tags:           []string{"Core_Operation"},
		ChallengeCode:  constants.ChallengeListUpgradeVersion,
		Doc:            "List operation with upgrade id",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Operation{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "upgrade-plan-id",
				Description:   "upgrade-plan-id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.Operation{},
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
		Path:           "/operation/{operation-id}/continue",
		HTTPMethod:     http.MethodPut,
		Handler:        operation.ContinueOperation,
		Tags:           []string{"Core_Operation"},
		ChallengeCode:  constants.ChallengeCodeContinueTask,
		Doc:            "List cluster operation",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Operation{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "operation-id",
				Description:   "operation id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.Operation{},
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
}
