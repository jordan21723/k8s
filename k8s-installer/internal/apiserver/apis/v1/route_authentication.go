package v1

import (
	"k8s-installer/internal/apiserver/controller/user"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var routeAuthentication = []Route{
	{
		Path:          "/login",
		HTTPMethod:    http.MethodPost,
		Handler:       user.Login,
		ChallengeCode: 0, // no need, because we always skip login api authorization, eggs or chicken which comes first ?
		Doc:           "Login",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: schema.User{},
		WriteDataModel: struct {
			Token       string      `json:"token"`
			ExpiredDate string      `json:"expired_date"`
			Permission  uint64      `json:"permission"`
			User        schema.User `json:"user"`
		}{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				struct {
					Token       string      `json:"token"`
					ExpiredDate string      `json:"expired_date"`
					Permission  uint64      `json:"permission"`
					User        schema.User `json:"user"`
				}{},
			},
			{
				http.StatusBadRequest,
				"bad request",
				schema.HttpErrorResult{},
			},
			{
				http.StatusUnauthorized,
				"Unauthorized",
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
		Handler:        user.CreateOrUpdateUser,
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "Create User",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.User{},
		WriteDataModel: schema.User{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.User{},
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
		Path:           "/{username}",
		HTTPMethod:     http.MethodDelete,
		Handler:        user.DeleteUser,
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "Delete User",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "username",
				Description:   "username",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: nil,
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
		Path:           "/{username}",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetUser,
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "Get User Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.User{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "username",
				Description:   "username",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.User{},
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
		Path:          "/{username}",
		HTTPMethod:    http.MethodPut,
		Handler:       user.CreateOrUpdateUser,
		ChallengeCode: constants.ChallengeCodeManageUser,
		Doc:           "Update User Info",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Id       string `json:"id" description:"do not input auto generator"`
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
			Roles    []struct {
				Id string `json:"id" description:"do not input auto generator"`
			} `json:"roles,omitempty" description:"contain one or more roles"`
		}{},
		WriteDataModel: schema.User{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "username",
				Description:   "username",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.User{},
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
		Path:           "/users",
		HTTPMethod:     http.MethodGet,
		Handler:        user.ListUser,
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "List User Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.User{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.User{},
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

var roleManage = []Route{

	{
		Path:          "/",
		HTTPMethod:    http.MethodPost,
		Handler:       user.CreateOrUpdateRole,
		ChallengeCode: constants.ChallengeCodeManageRole,
		Doc:           "Create Role",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Name     string `json:"name" validate:"required"`
			Function uint64 `json:"function" validate:"required" description:"function code represent capability of the role"`
		}{},
		WriteDataModel: schema.Role{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.Role{},
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
		Path:           "/{role-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        user.DeleteRole,
		ChallengeCode:  constants.ChallengeCodeManageRole,
		Doc:            "Delete Role",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: nil,
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
		Path:           "/{role-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetRole,
		ChallengeCode:  constants.ChallengeCodeManageRole,
		Doc:            "Get Role Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Role{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.Role{},
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
		Path:          "/{role-id}",
		HTTPMethod:    http.MethodPut,
		Handler:       user.CreateOrUpdateRole,
		ChallengeCode: constants.ChallengeCodeManageRole,
		Doc:           "Update Role Info",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Name     string `json:"name" validate:"required"`
			Function uint64 `json:"function" validate:"required" description:"function code represent capability of the role"`
		}{},
		WriteDataModel: schema.Role{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: schema.Role{},
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
		Path:           "/roles",
		HTTPMethod:     http.MethodGet,
		Handler:        user.ListRole,
		ChallengeCode:  constants.ChallengeCodeManageRole,
		Doc:            "List Role Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Role{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.Role{},
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
		Path:           "/roles_list",
		HTTPMethod:     http.MethodGet,
		Handler:        user.ListRole,
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "Roles for user creation,ensure user can be created when no role management permission is grant",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Role{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				HTTPStatus:  http.StatusOK,
				Message:     "OK",
				ReturnModel: []schema.Role{},
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
