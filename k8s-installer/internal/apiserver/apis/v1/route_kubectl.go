package v1

import (
	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/kubectl"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"
)

var kubectlRunCommandManage = []Route{
	{
		Path:           "/kubectl-exec",
		HTTPMethod:     http.MethodPost,
		Handler:        kubectl.KubectlExec,
		ChallengeCode:  constants.NoAuthRequire,
		Tags:           []string{"Core_KubectlExec"},
		Doc:            "Run kubectl exec",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.KubectlExec{},
		WriteDataModel: []schema.KubectlExecResult{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]schema.KubectlExecResult{},
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
