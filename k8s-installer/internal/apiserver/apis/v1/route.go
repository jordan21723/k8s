package v1

import (
	"k8s-installer/pkg/server/runtime"
	"k8s-installer/schema"

	"github.com/emicklei/go-restful"
)

type Path struct {
	GroupVersion runtime.GroupVersion
	Routes       []Route
}

/*
about authorization
simple scenario
ChallengeCode is suggest to use single bit e.g. ob 0010 =2
so it would be simply to verified it against sum up user role permission
[ChallengeCode] ob 0010 & [total user role permission] ob 1111 = 0010 (allow access)
[ChallengeCode] ob 0010 & [total user role permission] ob 1101 = 0000 (deny)

complex scenario
[ChallengeCode] ob 0110 & [total user role permission] ob 1111 = 0110 (allow access)
[ChallengeCode] ob 0110 & [total user role permission] ob 1011 = 0010 (deny)
*/

type Route struct {
	Path              string
	HTTPMethod        string
	Handler           func(request *restful.Request, response *restful.Response)
	ChallengeCode     uint64
	Doc               string
	PathParams        []schema.Parameter
	QueryParams       []schema.Parameter
	Filter            []restful.FilterFunction
	MetaData          string
	ReadDataModel     interface{}
	WriteDataModel    interface{}
	ReturnDefinitions []DocReturnDefinition
	Tags              []string
}

type DocReturnDefinition struct {
	HTTPStatus  int
	Message     string
	ReturnModel interface{}
}

var Paths = []Path{
	{
		GroupVersion: runtime.GroupVersion{Group: "core", Version: "v1"},
		Routes: appendsRoutes(
			nodeManageV2,
			regionManage,
			clusterManageV2,
			licenseV2Manage,
			operationV2Manager,
			upgradeV2Manage,
			AuthenticationV2Routes,
			roleManageV2,
			dansManage,
			kubectlRunCommandManage,
		),
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "", Version: ""},
		Routes: appendsRoutes(
			systemConfigRoute,
		),
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "cluster", Version: "v1"},
		Routes:       clusterManage,
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "user", Version: "v1"},
		Routes:       routeAuthentication,
	},
	// front-end has not used it, now Removed.
	// {
	// 	GroupVersion: runtime.GroupVersion{Group: "node", Version: "v1"},
	// 	Routes:       nodeManage,
	// },
	{
		GroupVersion: runtime.GroupVersion{Group: "role", Version: "v1"},
		Routes:       roleManage,
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "operation", Version: "v1"},
		Routes:       operationManager,
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "addons", Version: "v1"},
		Routes:       addonsManage,
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "upgrade", Version: "v1"},
		Routes:       upgradeManage,
	},
	{
		GroupVersion: runtime.GroupVersion{Group: "license", Version: "v1"},
		Routes:       licenseManage,
	},
}

func appendsRoutes(s []Route, elms ...[]Route) []Route {
	for _, v := range elms {
		s = append(s, v...)
	}
	return s
}
