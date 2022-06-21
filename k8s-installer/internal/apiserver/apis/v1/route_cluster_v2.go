package v1

import (
	"k8s-installer/internal/apiserver/controller/addons"
	"k8s-installer/internal/apiserver/controller/cluster"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var clusterManageV2 = []Route{
	{
		Path:           "/clusters",
		HTTPMethod:     http.MethodPost,
		Handler:        cluster.CreateCluster,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeCreateCluster,
		Doc:            "Create cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.Cluster{},
		WriteDataModel: schema.Cluster{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.Cluster{},
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
		Path:          "/clusters/{cluster-id}/update-status",
		HTTPMethod:    http.MethodPatch,
		Handler:       cluster.UpdateClusterStatus,
		Tags:          []string{"Core_Clusters"},
		ChallengeCode: constants.ChallengeCodeCreateCluster,
		Doc:           "update cluster status",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Status string `json:"status"`
		}{},
		WriteDataModel: struct {
			Status string `json:"status"`
		}{},
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
				ReturnModel: schema.Cluster{},
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
		Path:          "/clusters/{cluster-id}/update-cluster-installer",
		HTTPMethod:    http.MethodPatch,
		Handler:       cluster.UpdateClusterInstaller,
		Tags:          []string{"Core_Clusters"},
		ChallengeCode: constants.ChallengeCodeCreateCluster,
		Doc:           "update cluster installer",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			ClusterInstaller string `json:"cluster_installer"`
		}{},
		WriteDataModel: struct {
			ClusterInstaller string `json:"cluster_installer"`
		}{},
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
				ReturnModel: schema.Cluster{},
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
		Path:           "/clusters/{cluster-id}/recreate",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.ReCreateCluster,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeCreateCluster,
		Doc:            "Re-create cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Cluster{},
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
				ReturnModel: schema.Cluster{},
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
		Path:           "/clusters/{cluster-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        cluster.DeleteClusterV2,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeDeleteCluster,
		Doc:            "Delete cluster by given cluster id",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  []string{},
		WriteDataModel: nil,
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
		QueryParams: []schema.Parameter{
			{
				Required:      false,
				DataFormat:    "bool",
				DefaultValue:  "false",
				Name:          "gracefully_delete",
				Description:   "gracefully_delete delete namespace",
				DataType:      "bool",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "ok",
				ReturnModel: nil,
			},
			{
				http.StatusPreconditionFailed,
				"Not allowed to delete cluster",
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
		Path:           "/clusters/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.GetClusterInfo,
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Tags:           []string{"Core_Clusters"},
		Doc:            "Get cluster info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Cluster{},
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
				Message:     "ok",
				ReturnModel: schema.Cluster{},
			},
			{
				http.StatusBadRequest,
				"Unable to find cluster",
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
		Path:           "/clusters",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.ListCluster,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "List cluster info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Cluster{},
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
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "ok",
				ReturnModel: []schema.Cluster{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		//Path:           "/status/running",
		// TODO: move this to list clusters api as a query parameter
		Path:           "/clusters/running",
		HTTPMethod:     http.MethodGet,
		Tags:           []string{"Core_Clusters"},
		Handler:        cluster.ListRunningCluster,
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "List all running cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Cluster{},
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
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "ok",
				ReturnModel: []schema.Cluster{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/clusters/{cluster-id}/nodes",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.AddNodeToCluster,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeAddNodeToCluster,
		Doc:            "add node to cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  []schema.ClusterNode{},
		WriteDataModel: []schema.ClusterNode{},
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
				Message:     "ok",
				ReturnModel: []schema.Cluster{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
			{
				http.StatusBadRequest,
				"Unable to add node to cluster",
				schema.HttpErrorResult{},
			},
			{
				http.StatusPreconditionFailed,
				"Unable to add node to cluster due to pre-check failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/clusters/{cluster-id}/nodes",
		HTTPMethod:     http.MethodDelete,
		Handler:        cluster.RemoveNodeFromCluster,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeRemoveNodeFromCluster,
		Doc:            "remove node from cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  []schema.ClusterNode{},
		WriteDataModel: nil,
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
				HTTPStatus: http.StatusOK,
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
			{
				http.StatusBadRequest,
				"Unable to remove node from cluster with given cluster id",
				schema.HttpErrorResult{},
			},
			{
				http.StatusPreconditionFailed,
				"Unable to add node to cluster due to pre-check failed",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		//Path:           "/list_cluster_nodes/{cluster-id}",
		Path:           "/clusters/{cluster-id}/nodes",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.ListClusterNodes,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "list cluster nodes",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.NodeInformation{},
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
				Message:     "ok",
				ReturnModel: []schema.Cluster{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
			{
				http.StatusBadRequest,
				"Unable to list node with given cluster id",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/clusters/template",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.GetTemplate,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeConfigPlatform,
		Doc:            "get cluster template",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Cluster{},
		PathParams:     nil,
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "ok",
				ReturnModel: schema.Cluster{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
			{
				http.StatusBadRequest,
				"Unable to list node with given cluster id",
				schema.HttpErrorResult{},
			},
		},
	},

	{
		Path:           "/clusters/template",
		HTTPMethod:     http.MethodPost,
		Handler:        cluster.CreateTemplate,
		Tags:           []string{"Core_Clusters"},
		ChallengeCode:  constants.ChallengeCodeCreateCluster,
		Doc:            "create cluster template",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.ClusterTemplate{},
		WriteDataModel: schema.ClusterTemplate{},
		PathParams:     nil,
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "ok",
				ReturnModel: schema.ClusterTemplate{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
			{
				http.StatusBadRequest,
				"Unable to list node with given cluster id",
				schema.HttpErrorResult{},
			},
		},
	},
	{
		Path:           "/clusters/{cluster-id}/addons",
		HTTPMethod:     http.MethodGet,
		Handler:        addons.ListClusterAddons,
		Tags:           []string{"Core_Clusters"},
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
		Path:           "/clusters/{cluster-id}/addons",
		HTTPMethod:     http.MethodPut,
		Handler:        addons.ManageClusterAddons,
		Tags:           []string{"Core_Clusters"},
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
