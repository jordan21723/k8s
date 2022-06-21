package v1

import (
	"k8s-installer/internal/apiserver/controller/user"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var AuthenticationV2Routes = []Route{
	{
		Path:       "/send-message",
		HTTPMethod: http.MethodPost,
		Handler:    user.SendMesssageCode,
		//Handler:       user.CheckLoginType,
		Tags:          []string{"Core_Authentication"},
		ChallengeCode: 0, // no need, because we always skip login api authorization, eggs or chicken which comes first ?
		Doc:           "SendMesssageCode",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: schema.User{},
		QueryParams: []schema.Parameter{
			{
				Required:      true,
				Name:          "type",
				Description:   "message code type, installer=103, ks=103",
				DataType:      "string",
				AllowMultiple: false,
				DataFormat:    "string",
			},
		},
		WriteDataModel: struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				schema.SendMesssageResp{},
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
		Path:       "/text-code-validation",
		HTTPMethod: http.MethodPost,
		Handler:    user.ValidateMesssageCode,
		//Handler:       user.CheckLoginType,
		Tags:           []string{"Core_Authentication"},
		ChallengeCode:  0, // no need, because we always skip login api authorization, eggs or chicken which comes first ?
		Doc:            "ValidateMesssageCode",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.ValidateMesssageCode{},
		WriteDataModel: schema.ValidateMesssageResp{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				schema.ValidateMesssageResp{},
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
		Path:       "/login",
		HTTPMethod: http.MethodPost,
		Handler:    user.Login,
		//Handler:       user.CheckLoginType,
		Tags:          []string{"Core_Authentication"},
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
		Path:          "/users",
		HTTPMethod:    http.MethodPost,
		Handler:       user.CreateOrUpdateUser,
		Tags:          []string{"Core_User"},
		ChallengeCode: constants.ChallengeCodeManageUser,
		Doc:           "Create User",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Username  string        `json:"username" validate:"required"`
			Password  string        `json:"password" validate:"required"`
			Roles     []schema.Role `json:"roles,omitempty" description:"contain one or more roles"`
			Phone     string        `json:"phone,omitempty" description:"user phone number"`
			KsAccount string        `json:"ks_account,omitempty" description:"kubespere user account"`
		}{},
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
		Path:           "/users/{username}",
		HTTPMethod:     http.MethodDelete,
		Tags:           []string{"Core_User"},
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
		Path:           "/users/{username}",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetUser,
		Tags:           []string{"Core_User"},
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
		Path:          "/users/{username}",
		HTTPMethod:    http.MethodPut,
		Handler:       user.CreateOrUpdateUser,
		Tags:          []string{"Core_User"},
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
			Phone     string `json:"phone,omitempty" description:"user phone number"`
			KsAccount string `json:"ks_account,omitempty" description:"kubespere user account"`
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
		Tags:           []string{"Core_User"},
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
	{
		Path:           "/users/{username}/globalroles",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetUserGlobalRules,
		Tags:           []string{"Core_User"},
		ChallengeCode:  constants.ChallengeCodeManageUser,
		Doc:            "Get User Global Rules Info, only used in caas auth type",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.CaaSGlobalRoles{},
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
				ReturnModel: []schema.CaaSGlobalRoles{},
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

var roleManageV2 = []Route{

	{
		Path:          "/roles",
		HTTPMethod:    http.MethodPost,
		Handler:       user.CreateOrUpdateRole,
		Tags:          []string{"Core_Role"},
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
		Path:           "/roles/{role-id}",
		HTTPMethod:     http.MethodDelete,
		Handler:        user.DeleteRole,
		Tags:           []string{"Core_Role"},
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
		Path:           "/roles/{role-id}",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetRole,
		Tags:           []string{"Core_Role"},
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
		Path:           "/roles/{role-id}",
		HTTPMethod:     http.MethodPut,
		Handler:        user.CreateOrUpdateRole,
		Tags:           []string{"Core_Role"},
		ChallengeCode:  constants.ChallengeCodeManageRole,
		Doc:            "Update Role Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  schema.Role{},
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
		Tags:           []string{"Core_Role"},
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
	// 重复 url ?
	// {
	// 	Path:           "/roles_list",
	// 	HTTPMethod:     http.MethodGet,
	// 	Handler:        user.ListRole,
	// 	ChallengeCode:  constants.ChallengeCodeManageUser,
	// 	Doc:            "Roles for user creation,ensure user can be created when no role management permission is grant",
	// 	MetaData:       restfulSpec.KeyOpenAPITags,
	// 	ReadDataModel:  nil,
	// 	WriteDataModel: []schema.Role{},
	// 	ReturnDefinitions: []DocReturnDefinition{
	// 		{
	// 			HTTPStatus:  http.StatusOK,
	// 			Message:     "OK",
	// 			ReturnModel: []schema.Role{},
	// 		},
	// 		{
	// 			http.StatusBadRequest,
	// 			"bad request",
	// 			schema.HttpErrorResult{},
	// 		},
	// 		{
	// 			http.StatusInternalServerError,
	// 			"Internal Server Error",
	// 			schema.HttpErrorResult{},
	// 		},
	// 	},
	// },
}

var systemConfigRoute = []Route{
	{
		Path:           "/authtype",
		HTTPMethod:     http.MethodGet,
		Handler:        user.GetAuthType,
		Tags:           []string{"System"},
		ChallengeCode:  0, // no need, because we always skip login api authorization, eggs or chicken which comes first ?
		Doc:            "Get Core_Authentication_Type",
		MetaData:       restfulSpec.KeyOpenAPITags,
		WriteDataModel: schema.SystemAuthType{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				schema.SystemAuthType{},
			},
			{
				http.StatusInternalServerError,
				"Internal Server Error",
				schema.HttpErrorResult{},
			},
		},
	},
}
