package v1

import (
	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/addons"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"
)

var addonsManage = []Route{
	{
		Path:           "/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        addons.ListClusterAddons,
		ChallengeCode:  constants.ChallengeCodeListNode,
		Doc:            "List Addons",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Addons{},
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
				ReturnModel: schema.Addons{},
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
		Path:           "/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        addons.ManageClusterAddons,
		ChallengeCode:  constants.ChallengeCodeListNode,
		Doc:            "List Addons",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.Addons{},
		WriteDataModel: schema.Addons{},
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
				ReturnModel: schema.Addons{},
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
				"Precondition Check Failed",
				schema.HttpErrorResult{},
			},
		},
	},
}
