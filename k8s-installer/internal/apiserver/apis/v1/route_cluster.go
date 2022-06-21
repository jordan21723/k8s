package v1

import (
	"net/http"

	"k8s-installer/internal/apiserver/controller/cluster"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var clusterManage = []Route{
	{
		Path:           "/",
		HTTPMethod:     http.MethodPost,
		Handler:        cluster.CreateCluster,
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
		Path:           "/recreate/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.ReCreateCluster,
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
		Path:           "/{cluster-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        cluster.DeleteCluster,
		ChallengeCode:  constants.ChallengeCodeDeleteCluster,
		Doc:            "Delete cluster by given cluster id",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
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
		Path:           "/gracefully_delete/{cluster-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        cluster.DeleteCluster,
		ChallengeCode:  constants.ChallengeCodeDeleteCluster,
		Doc:            "Delete cluster by gracefully",
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
		Path:           "/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.GetClusterInfo,
		ChallengeCode:  constants.ChallengeCodeListCluster,
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
		Path:           "/",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.ListCluster,
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "List cluster info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.ClusterApi{},
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
		Path:           "/status/running",
		HTTPMethod:     http.MethodGet,
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
		Path:           "/ks/host",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.ListKsHostCluster,
		ChallengeCode:  constants.ChallengeCodeListCluster,
		Doc:            "List all ks host cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Cluster{},
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
		Path:           "/add_node/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.AddNodeToCluster,
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
		Path:           "/remove_node/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.RemoveNodeFromCluster,
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
		Path:           "/remove_node/db_only/{cluster-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        cluster.RemoveNodeFromClusterDBOnly,
		ChallengeCode:  constants.ChallengeCodeRemoveNodeFromCluster,
		Doc:            "force remove node from cluster in database without touch actual k8s installer",
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
		Path:           "/list_cluster_nodes/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.ListClusterNodes,
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
		Path:           "/template",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.GetTemplate,
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
		Path:           "/template",
		HTTPMethod:     http.MethodPost,
		Handler:        cluster.CreateTemplate,
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
		Path:          "/backup/{cluster-id}",
		HTTPMethod:    http.MethodPost,
		Handler:       cluster.BackUpCreate,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Create Backup",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: schema.Backup{},
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
				Message:    "OK",
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

	{
		Path:          "/backup/{cluster-id}/{backup-name}",
		HTTPMethod:    http.MethodDelete,
		Handler:       cluster.BackUpDelete,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Delete Backup",
		MetaData:      restfulSpec.KeyOpenAPITags,
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
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "backup-name",
				Description:   "backup name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
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

	{
		Path:           "/backup/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.BackUpList,
		Tags:           []string{"Backup_Restore"},
		ChallengeCode:  constants.ChallengeCodeBackupAndRestore,
		Doc:            "Cluster List Backup",
		MetaData:       restfulSpec.KeyOpenAPITags,
		WriteDataModel: schema.BackupList{},
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
				ReturnModel: schema.BackupList{},
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

	{
		Path:          "/restore/{cluster-id}/{backup-name}",
		HTTPMethod:    http.MethodPost,
		Handler:       cluster.RestoreCreate,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Create Backup",
		MetaData:      restfulSpec.KeyOpenAPITags,
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
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "backup-name",
				Description:   "restore cluster from backup file name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
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

	{
		Path:           "/restore/{cluster-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.RestoreList,
		Tags:           []string{"Backup_Restore"},
		ChallengeCode:  constants.ChallengeCodeBackupAndRestore,
		Doc:            "Cluster List Restore",
		MetaData:       restfulSpec.KeyOpenAPITags,
		WriteDataModel: schema.RestoreList{},
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
				ReturnModel: schema.RestoreList{},
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

	{
		Path:          "/restore/{cluster-id}/{restore-name}",
		HTTPMethod:    http.MethodDelete,
		Handler:       cluster.RestoreDelete,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Delete Restore",
		MetaData:      restfulSpec.KeyOpenAPITags,
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
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "restore-name",
				Description:   "restore name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
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

	{
		Path:          "/backupregular/{cluster-id}",
		HTTPMethod:    http.MethodPost,
		Handler:       cluster.BackupRegularCreate,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Create Regular Backup",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: schema.BackupRegular{},
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
				ReturnModel: schema.BackupRegular{},
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

	{
		Path:          "/backupregular/{cluster-id}/{backup-regular-name}",
		HTTPMethod:    http.MethodDelete,
		Handler:       cluster.BackupRegularDelete,
		Tags:          []string{"Backup_Restore"},
		ChallengeCode: constants.ChallengeCodeBackupAndRestore,
		Doc:           "Cluster Delete Regular Backup",
		MetaData:      restfulSpec.KeyOpenAPITags,
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
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "backup-regular-name",
				Description:   "backupregular name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
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
	{
		Path:           "/backupregular/{cluster-id}/{backup-regular-name}",
		HTTPMethod:     http.MethodGet,
		Handler:        cluster.BackupRegularGet,
		Tags:           []string{"Backup_Restore"},
		ChallengeCode:  constants.ChallengeCodeBackupAndRestore,
		Doc:            "Cluster Get Regular Backup",
		MetaData:       restfulSpec.KeyOpenAPITags,
		WriteDataModel: schema.BackupRegularlyDetail{},
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
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "backup-regular-name",
				Description:   "backupregular name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus: http.StatusOK,
				Message:    "OK",
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
	{
		Path:           "/mock/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.MockOrUnMockCluster,
		Tags:           []string{"Mock_Cluster"},
		ChallengeCode:  constants.ChallengeCodeCreateCluster,
		Doc:            "mock a cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
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
				Message:    "OK",
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
		Path:           "/unmock/{cluster-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        cluster.MockOrUnMockCluster,
		Tags:           []string{"Mock_Cluster"},
		ChallengeCode:  constants.ChallengeCodeCreateCluster,
		Doc:            "mock a cluster",
		MetaData:       restfulSpec.KeyOpenAPITags,
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
				Message:    "OK",
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
