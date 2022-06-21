package v1

import (
	"github.com/emicklei/go-restful"

	"k8s-installer/pkg/server/params"
	"k8s-installer/pkg/server/runtime"
)

const GroupName = "resources.io"

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebServiceWithStaticRoute)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebServiceWithStaticRoute(c *restful.Container) error {
	for _, path := range Paths {
		webservice := runtime.NewWebService(path.GroupVersion)
		for _, route := range path.Routes {
			routeBuilder := webservice.Method(route.HTTPMethod).
				Path(route.Path).
				To(route.Handler).
				Doc(route.Doc).
				Metadata(route.MetaData, route.Tags).
				Writes(route.WriteDataModel)
			if route.ReadDataModel != nil {
				routeBuilder = routeBuilder.Reads(route.ReadDataModel)
			}
			routeBuilder = params.PathParameterBuilder(routeBuilder, route.PathParams)
			routeBuilder = params.QueryParameterBuilder(routeBuilder, route.QueryParams)
			for _, returnDef := range route.ReturnDefinitions {
				routeBuilder = routeBuilder.Returns(returnDef.HTTPStatus, returnDef.Message, returnDef.ReturnModel)
			}
			//Add filter in router level
			for _, filter := range route.Filter {
				routeBuilder.Filter(filter)
			}

			webservice.Route(routeBuilder)
		}
		c.Add(webservice)
	}
	return nil
}
