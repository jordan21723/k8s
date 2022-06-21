package v1

import (
	"net/http"

	"k8s-installer/internal/apiserver/controller/node"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var nodeManageV2 = []Route{
	{
		Path:          "/nodes",
		HTTPMethod:    http.MethodGet,
		Handler:       node.ListNode,
		ChallengeCode: constants.ChallengeCodeListNode,
		Tags:          []string{"Core_Node"},
		Doc:           "List Node Info",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: nil,
		QueryParams: []schema.Parameter{
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
	{
		Path:          "/nodes/usable",
		HTTPMethod:    http.MethodGet,
		Handler:       node.ListUsableNode,
		ChallengeCode: constants.ChallengeCodeCreateCluster,
		Tags:          []string{"Core_Node"},
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
	{
		Path:           "/nodes/{node-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        node.GetNode,
		ChallengeCode:  constants.ChallengeCodeListNode,
		Tags:           []string{"Core_Node"},
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
		Path:           "/nodes/{node-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        node.DeleteNode,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Tags:           []string{"Core_Node"},
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
	{
		Path:           "/nodes/{node-id}/disable",
		HTTPMethod:     http.MethodPatch,
		Handler:        node.DisableNode,
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
			{
				http.StatusPreconditionFailed,
				"Precondition Failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/nodes/{node-id}/enable",
		HTTPMethod:     http.MethodPatch,
		Handler:        node.EnableNode,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Tags:           []string{"Core_Node"},
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
		Path:           "/ssh",
		HTTPMethod:     http.MethodGet,
		Handler:        node.NodeSSH,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Tags:           []string{"Core_Node"},
		Doc:            "Connector node with ssh protocol",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		QueryParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "token",
				Description:   "auth token",
				DataType:      "string",
				AllowMultiple: false,
			},
			{
				Required:      true,
				DataFormat:    "int",
				DefaultValue:  "150",
				Name:          "cols",
				Description:   "terminal cols",
				DataType:      "int",
				AllowMultiple: false,
			},
			{
				Required:      true,
				DataFormat:    "int",
				DefaultValue:  "35",
				Name:          "rows",
				Description:   "terminal rows",
				DataType:      "int",
				AllowMultiple: false,
			},
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "msg",
				Description:   "node ip,username,password,port,use base64 encode",
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
		Path:           "/ssh/rsakey",
		HTTPMethod:     http.MethodGet,
		Handler:        node.GetSSHRSAKey,
		ChallengeCode:  constants.ChallengeCodeEnableOrDisableNode,
		Tags:           []string{"Core_Node"},
		Doc:            "Get rsa public key",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.SSHRSAkey{},
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
