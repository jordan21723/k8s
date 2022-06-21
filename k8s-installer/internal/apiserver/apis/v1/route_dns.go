package v1

import (
	"k8s-installer/schema"
	"net/http"

	"k8s-installer/internal/apiserver/controller/dns"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/coredns"

	restfulSpec "github.com/emicklei/go-restful-openapi"
)

var dansManage = []Route{
	{
		Path:           "/domain/sub-domain",
		HTTPMethod:     http.MethodGet,
		Handler:        dns.ListAllSubDomain,
		ChallengeCode:  constants.NoAuthRequire,
		Tags:           []string{"Core_Dns"},
		Doc:            "List all sub domain and it's resolve",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: []coredns.DNSDomain{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]coredns.DNSDomain{},
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
		Path:          "/domain/{domain}/sub-domain",
		HTTPMethod:    http.MethodGet,
		Handler:       dns.ListSubDomainOfTLd,
		ChallengeCode: constants.NoAuthRequire,
		Tags:          []string{"Core_Dns"},
		Doc:           "List all sub domain and it's resolve",
		MetaData:      restfulSpec.KeyOpenAPITags,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "top domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReadDataModel:  nil,
		WriteDataModel: []coredns.DNSDomain{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]coredns.DNSDomain{},
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
		Path:          "/domain/{domain}/sub-domain",
		HTTPMethod:    http.MethodPost,
		Handler:       dns.CreateSubDomain,
		ChallengeCode: constants.ChallengeCodeManageDNS,
		Tags:          []string{"Core_Dns"},
		Doc:           "Create sub domain and it's resolve of a top level domain",
		MetaData:      restfulSpec.KeyOpenAPITags,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "top level domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReadDataModel:  coredns.DNSDomainUpdateDocSchema{},
		WriteDataModel: coredns.DNSDomainUpdateDocSchema{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]coredns.DNSDomain{},
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
		Path:          "/domain/{tld-domain}/sub-domain/{domain}",
		HTTPMethod:    http.MethodPut,
		Handler:       dns.UpdateSubDomain,
		ChallengeCode: constants.ChallengeCodeManageDNS,
		Tags:          []string{"Core_Dns"},
		Doc:           "Update sub domain",
		MetaData:      restfulSpec.KeyOpenAPITags,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "tld-domain",
				Description:   "top level domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReadDataModel: struct {
			DomainResolve []coredns.DNSDomainUpdateDocSchema `json:"domain_resolve"  validate:"required" description:"ip address list of this domain resolve to"`
			Description   string                             `json:"description,omitempty"`
		}{},
		WriteDataModel: coredns.DNSDomain{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]coredns.DNSDomain{},
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
		Path:           "/domain/top-level-domain",
		HTTPMethod:     http.MethodGet,
		Handler:        dns.ListTopLevelDomains,
		ChallengeCode:  constants.ChallengeCodeManageDNS,
		Tags:           []string{"Core_Dns"},
		Doc:            "List top level domain",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: coredns.TopLevelDomain{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				[]coredns.TopLevelDomain{},
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
		Path:           "/domain/top-level-domain/{domain}",
		HTTPMethod:     http.MethodGet,
		Handler:        dns.GetTopLevelDomain,
		ChallengeCode:  constants.ChallengeCodeManageDNS,
		Tags:           []string{"Core_Dns"},
		Doc:            "Get top level domain detail",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: coredns.TopLevelDomain{},
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				coredns.TopLevelDomain{},
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
		Path:           "/domain/top-level-domain/{domain}",
		HTTPMethod:     http.MethodDelete,
		Handler:        dns.DeleteTopLevelDomain,
		ChallengeCode:  constants.ChallengeCodeManageDNS,
		Tags:           []string{"Core_Dns"},
		Doc:            "Delete top level domain",
		MetaData:       restfulSpec.KeyOpenAPITags,
		ReadDataModel:  nil,
		WriteDataModel: nil,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				nil,
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
		Path:          "/domain/top-level-domain",
		HTTPMethod:    http.MethodPost,
		Handler:       dns.CreateTopLevelDomain,
		ChallengeCode: constants.ChallengeCodeManageDNS,
		Tags:          []string{"Core_Dns"},
		Doc:           "create top level domain",
		MetaData:      restfulSpec.KeyOpenAPITags,
		ReadDataModel: struct {
			Domain      string `json:"domain" validate:"required,fqdn" description:"domain name unique"`
			Description string `json:"description,omitempty"`
		}{},
		WriteDataModel: coredns.TopLevelDomain{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				coredns.TopLevelDomain{},
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
		Path:          "/domain/top-level-domain/{domain}",
		HTTPMethod:    http.MethodPut,
		Handler:       dns.UpdateTopLevelDomain,
		ChallengeCode: constants.ChallengeCodeManageDNS,
		Tags:          []string{"Core_Dns"},
		Doc:           "update dns domain",
		MetaData:      restfulSpec.KeyOpenAPITags,
		PathParams: []schema.Parameter{
			{
				Required:      true,
				DataFormat:    "string",
				DefaultValue:  "",
				Name:          "domain",
				Description:   "domain name",
				DataType:      "string",
				AllowMultiple: false,
			},
		},
		ReadDataModel: struct {
			Description string `json:"description,omitempty"`
		}{},
		WriteDataModel: struct {
			Description string `json:"description,omitempty"`
		}{},
		ReturnDefinitions: []DocReturnDefinition{
			{
				http.StatusOK,
				"OK",
				struct {
					Description string `json:"description,omitempty"`
				}{},
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
