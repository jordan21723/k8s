package v1

import (
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/node"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

var nodeManage = []Route{
	{
		Path:          "/",
		HTTPMethod:    http.MethodGet,
		Handler:       node.ListNode,
		ChallengeCode: constants.ChallengeCodeListNode,
		Doc:           "List Node Info",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: nil,
		QueryParams: []schema.Parameter{
			{
				Required:      false,
				DataFormat:    "string",
				DefaultValue:  "kubeadm",
				Name:          "cluster_installer",
				Description:   "cluster installer",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		WriteDataModel: []schema.NodeInformation{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.NodeInformation{},
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
	/*
		{
			Path:          "/usable",
			HTTPMethod:    http.MethodGet,
			Handler:       node.ListUsableNode,
			ChallengeCode: constants.ChallengeCodeCreateCluster,
			Doc:           "List Node That Can Be Installed With Cluster",
			MetaData:      restfulSpec.KeyOpenAPITags,
			ReadDataModel: nil,
			QueryParams: []schema.Parameter{
				{
					Required:      true,
					DataFormat:    "int",
					DefaultValue:  "0",
					Name:          "page_index",
					Description:   "page index",
					DataType:      "int",
					AllowMultiple: false,
				},
				{
					Required:      true,
					DataFormat:    "int",
					DefaultValue:  "0",
					Name:          "page_size",
					Description:   "page size",
					DataType:      "int",
					AllowMultiple: false,
				},
				{
					Required:      false,
					DataFormat:    "string",
					DefaultValue:  "not-set",
					Name:          "region_id",
					Description:   "region id",
					DataType:      "string",
					AllowMultiple: false,
				},
			},
			WriteDataModel: struct {
				TotalPage int64                    `json:"total_page"`
				Nodes     []schema.NodeInformation `json:"nodes"`
			}{},
			ReturnDefinitions: []DocReturnDefinition{
				{
					HTTPStatus:  http.StatusOK,
					Message:     "OK",
					ReturnModel: []schema.NodeInformation{},
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
	*/
	{
		Path:           "/{node-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        node.GetNode,
		ChallengeCode:  constants.ChallengeCodeListNode,
		Doc:            "Get Node Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.NodeInformation{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "node-id",
				Description:   "node id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.NodeInformation{},
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
		Path:           "/{node-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        node.DeleteNode,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Doc:            "Delete node from db",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "node-id",
				Description:   "node id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.NodeInformation{},
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
				"Precondition Failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/disable/{node-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        node.EnableOrDisableNode,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Doc:            "disable node from db",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "node-id",
				Description:   "node id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.NodeInformation{},
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
				"Precondition Failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/enable/{node-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        node.EnableOrDisableNode,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Doc:            "Enable node from db",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "node-id",
				Description:   "node id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.NodeInformation{},
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
				"Precondition Failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/nodes/{node-id}/update-role",
		HTTPMethod:     http.MethodPatch,
		Handler:        node.UpdateNodeRole,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Tags:           []string{"Core_Node"},
		Doc:            "disable node from db",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "node-id",
				Description:   "node id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.NodeInformation{},
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
