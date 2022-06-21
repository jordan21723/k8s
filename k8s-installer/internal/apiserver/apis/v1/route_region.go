package v1

import (
	"net/http"

	restfulSpec "github.com/emicklei/go-restful-openapi"
	"k8s-installer/internal/apiserver/controller/region"
	"k8s-installer/pkg/constants"
	"k8s-installer/schema"
)

var regionManage = []Route{
	{
		Path:           "/regions",
		HTTPMethod:     http.MethodGet,
		Handler:        region.ListRegions,
		ChallengeCode:  constants.ChallengeCodeManageRegion,
		Tags:           []string{"Core_Region"},
		Doc:            "List Region Info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []schema.Region{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]schema.Region{},
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
		Path:           "/regions/{region}",
		HTTPMethod:     http.MethodGet,
		Handler:        region.GetRegion,
		ChallengeCode:  constants.ChallengeCodeManageRegion,
		Tags:           []string{"Core_Region"},
		Doc:            "Get Region info",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: schema.Region{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				schema.Region{},
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
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "region",
				Description:   "region id",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
	},
	/*
		{
			Path:           "/regions",
			HTTPMethod:     http.MethodPost,
			Handler:        region.CreateRegion,
			ChallengeCode:  constants.ChallengeCodeManageRegion,
			Tags:           []string{"Core_Region"},
			Doc:            "Create Region",
			MetaData:       restfulSpec.KeyOpenAPITags,
			ReadDataModel:  schema.Region{},
			WriteDataModel: schema.Region{},
			ReturnDefinitions: []DocReturnDefinition{
				{
					http.StatusCreated,
					"OK",
					schema.Region{},
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
		Path:           "/regions/{region}",
		HTTPMethod:     http.MethodDelete,
		Handler:        region.DeleteRegion,
		ChallengeCode:  constants.ChallengeCodeManageRegion,
		Tags:           []string{"Core_Region"},
		Doc:            "Delete region by given region id",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "region",
				Description:   "region id",
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
				"Not allowed to delete region",
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
			Path:           "/regions/{region}",
			HTTPMethod:     http.MethodPut,
			Handler:        region.UpdateRegion,
			ChallengeCode:  constants.ChallengeCodeManageRegion,
			Tags:           []string{"Core_Region"},
			Doc:            "Update region by given region id",
			MetaData:       restfulSpec.KeyOpenAPITags,
			ReadDataModel:  schema.Region{},
			WriteDataModel: schema.Region{},
			PathParams: []schema.Parameter{
				{
					Required:      true,
					DataFormat:    "string",
					DefaultValue:  "",
					Name:          "region",
					Description:   "region id",
					DataType:      "string",
					AllowMultiple: false,
				},
			},
			ReturnDefinitions: []DocReturnDefinition{
				{
					HTTPStatus:  http.StatusOK,
					Message:     "ok",
					ReturnModel: schema.Region{},
				},
				{
					http.StatusPreconditionFailed,
					"Not allowed to delete region",
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
}
