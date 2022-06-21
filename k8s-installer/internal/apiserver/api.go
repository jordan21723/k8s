package apiserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	apiUtil "k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	jwtPkg "k8s-installer/pkg/jwt"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/schema"

	apisv1 "k8s-installer/internal/apiserver/apis/v1"
	apiServerConfig "k8s-installer/pkg/config/api_server"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/server/runtime"
	"k8s-installer/pkg/util"

	"github.com/emicklei/go-restful"
	"k8s.io/apimachinery/pkg/util/sets"
)

var routeMapping = map[string]map[string]apisv1.Route{}

func InstallAPIs(container *restful.Container) {
	runtime.Must(apisv1.AddToContainer(container))
}
func CreateGoRestful() *restful.Container {
	wsContainer := restful.NewContainer()
	// license should validate before authentication
	cors := restful.CrossOriginResourceSharing{
		//ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"},
		AllowedMethods: []string{"POST, OPTIONS, GET, PUT"},
		CookiesAllowed: false,
		Container:      wsContainer}
	wsContainer.Filter(cors.Filter)
	wsContainer.Filter(LicenseValid)
	wsContainer.Filter(excludePathAuth([]string{"/api/user/v1/login", "/api/core/v1/login", "/api/authtype", "/api/core/v1/domain/sub-domain", "/api/core/v1/kubectl-exec", "api/core/v1/ssh", "/api/core/v1/licenses", "/api/license/v1/"}))
	InstallAPIs(wsContainer)
	setRouteMapping()
	return wsContainer
}

func Start(config apiServerConfig.ApiConfig, port uint32, handler http.Handler, stopChan <-chan struct{}, ctx context.Context) {
	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	server.Handler = handler
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	log.Debugf("server starts and listening at port %d", port)
	<-stopChan
	log.Debug("Stop signal received. server shutting down !!!")
	server.Shutdown(ctx)
}

func ResourceHandler(resourcePath string) http.Handler {
	if err := util.CreateDirIfNotExists(resourcePath); err != nil {
		log.Fatal(err)
	}
	log.Debugf("resource path set to: %s", resourcePath)
	return http.StripPrefix("/", http.FileServer(http.Dir(resourcePath)))
}

func setRouteMapping() {
	for _, path := range apisv1.Paths {
		baseRoute := runtime.ApiRootPath + "/" + path.GroupVersion.String()
		for _, route := range path.Routes {
			if _, exist := routeMapping[route.HTTPMethod]; !exist {
				routeMapping[route.HTTPMethod] = map[string]apisv1.Route{
					baseRoute + route.Path: route,
				}
			} else {
				routeMapping[route.HTTPMethod][baseRoute+route.Path] = route
			}
			//log.Infof("baseroute + route path is %s", baseRoute+route.Path)
			// routeMapping[route.HTTPMethod][baseRoute + route.Path] = route
			// 			routeMapping[route.HTTPMethod] = map[string]apisv1.Route{
			// 			baseRoute + route.Path: route,
			// 		}
		}
	}
}

func LicenseValid(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// Notice: 登录前应直接返回 license 的验证结果，如果 license 验证不通过应直接跳转 license 输入界面

	log.Debug(req.SelectedRoutePath())
	if req.SelectedRoutePath() == "/api/license/v1/" || req.SelectedRoutePath() == "/api/core/v1/licenses" {
		chain.ProcessFilter(req, resp)
		return
	}

	runtimeCache := cache.GetCurrentCache()
	license, err := runtimeCache.GetLicense()
	if err != nil {
		log.Errorf("get license error: %v", err)
		// TODO: use 402?
		// https://stackoverflow.com/questions/39221380/what-is-the-http-status-code-for-license-limit-reached
		apiUtil.ResponseErrorWithCode(resp, http.StatusUnauthorized, constants.ErrLicenseInvalid, err.Error())
		return
	}

	if license == nil {
		apiUtil.ResponseErrorWithCode(resp, http.StatusUnauthorized, constants.ErrLicenseInvalid, "license not found!")
	} else {
		systemInfo, err := fileutils.ParseLicense(license.License)
		if err != nil {
			log.Debug(err)
			apiUtil.ResponseErrorWithCode(resp, http.StatusUnauthorized, constants.ErrLicenseInvalid, systemInfo.ErrorDetail)
			return
		}
		if systemInfo.LicenseValid {
			chain.ProcessFilter(req, resp)
		} else {
			apiUtil.ResponseErrorWithCode(resp, http.StatusUnauthorized, constants.ErrLicenseInvalid, fmt.Sprintf("Invalid license due to '%s'", systemInfo.ErrorDetail))
			return
		}
	}
}

func excludePathAuth(alwaysAllowPaths []string) restful.FilterFunction {
	var prefixes []string
	paths := sets.NewString()
	for _, p := range alwaysAllowPaths {
		p = strings.TrimPrefix(p, "/")
		if len(p) == 0 {
			// matches "/"
			paths.Insert(p)
			continue
		}
		if strings.ContainsRune(p[:len(p)-1], '*') {
			panic(fmt.Errorf("only trailing * allowed in %q", p))
		}
		if strings.HasSuffix(p, "*") {
			prefixes = append(prefixes, p[:len(p)-1])
		} else {
			paths.Insert(p)
		}
	}

	return func(req *restful.Request, resp *restful.Response, fc *restful.FilterChain) {
		pth := strings.TrimPrefix(req.Request.URL.Path, "/")
		if paths.Has(pth) {
			fc.ProcessFilter(req, resp)
			return
		}
		for _, prefix := range prefixes {
			if strings.HasPrefix(pth, prefix) {
				fc.ProcessFilter(req, resp)
				return
			}
		}
		switch req.HeaderParameter(constants.AuthTypeHeader) {
		case constants.StandardAuth:
			standardAuth(req, resp, fc)
		case constants.CaaSAuth:
			caasAuth(req, resp, fc)
		default:
			standardAuth(req, resp, fc)
		}
	}
}

func standardAuth(req *restful.Request, resp *restful.Response, fc *restful.FilterChain) {
	var routeFound apisv1.Route
	var found bool
	if routeFound, found = routeMapping[req.Request.Method][req.SelectedRoutePath()]; found {
		if routeFound.ChallengeCode == constants.NoAuthRequire {
			fc.ProcessFilter(req, resp)
			return
		}

		headerToken := req.HeaderParameter("token")
		if headerToken == "" {
			if v := req.QueryParameter("token"); v != "" {
				tokenBytes, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					apiUtil.ResponseError(resp, http.StatusInternalServerError, "Failed to decode token")
					return
				}
				headerToken = string(tokenBytes)
			} else {
				cookie, err := req.Request.Cookie("token")
				if err != nil {
					apiUtil.ResponseError(resp, http.StatusBadRequest, "Unable to find property token in header or cookie")
					return
				} else {
					headerToken = cookie.Value
				}
			}
		}
		runtimeCache := cache.GetCurrentCache()
		config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)

		// going to do token validation
		signString, err := base64.StdEncoding.DecodeString(config.ApiServer.JWTSignString)
		if err != nil {
			apiUtil.ResponseError(resp, http.StatusInternalServerError, "Failed to decode jwt sign string")
			return
		}
		user := &schema.User{}
		if err := jwtPkg.ValidateJWT(headerToken, user, string(signString)); err != nil {
			// token validation failed
			apiUtil.ResponseError(resp, http.StatusUnauthorized, err.Error())
			return
		}
		// reload user information, always use server side data for security
		user, err = runtimeCache.GetUser(user.Username)
		if err != nil {
			apiUtil.ResponseError(resp, http.StatusInternalServerError, err.Error())
			return
		}
		if user == nil {
			errMsg := "Failed to get user info. Authentication filter error"
			apiUtil.ResponseError(resp, http.StatusForbidden, errMsg)
			return
		}

		rolesMapping, errRole := runtimeCache.GetRoleList()
		if errRole != nil {
			apiUtil.ResponseError(resp, http.StatusInternalServerError, "Failed to get role list from db during authentication check")
			log.Errorf("Failed to get role list from db during authentication check due to error: %s", errRole.Error())
			return
		}
		if rolesMapping == nil || len(rolesMapping) == 0 {
			apiUtil.ResponseError(resp, http.StatusInternalServerError, "Failed to get role list from db during authentication check")
			log.Errorf("Failed to get role list from db during authentication check due to role lis is nil or empty. Did role data proper config?")
			return
		}

		//log.Debugf("request select route path is %s", req.SelectedRoutePath())
		// try to find related route information

		userPermission := uint64(0)
		// combine user role`s permission together
		for _, role := range user.Roles {
			if roleFound, isFound := rolesMapping[role.Id]; isFound {
				userPermission |= roleFound.Function
			}
		}
		// do authorization
		if routeFound.ChallengeCode&userPermission == routeFound.ChallengeCode {
			// ob1101 & ob 0010 = 0 authorization failed
			// ob 0001 & ob 0001 = ob0001 authorization success
			// match function Challenge Code can process
			req.SetAttribute("run-by-user", user.Username)
			fc.ProcessFilter(req, resp)
		} else {
			apiUtil.ResponseError(resp, http.StatusForbidden, fmt.Sprintf("You do not have the proper permission to call api %s", req.SelectedRoutePath()))
			return
		}
	} else {
		// route not match , oops 404
		resp.WriteHeader(http.StatusNotFound)
		log.Errorf("Api request error: path %s is not found", req.Request.RequestURI)
		return
	}

}

func caasAuth(req *restful.Request, resp *restful.Response, fc *restful.FilterChain) {
	token, err := getToken(req)
	if err != nil {
		apiUtil.ResponseError(resp, http.StatusBadRequest, "Unable to find property token in header or cookie")
		return
	}
	user, err := getUserHeader(req)
	if err != nil {
		apiUtil.ResponseError(resp, http.StatusBadRequest, "Unable to find property token in header or cookie")
		return
	}

	log.Debug(req.Request.URL.Path)
	if req.Request.URL.Path == fmt.Sprintf("/api/core/v1/users/%s", user) ||
		req.Request.URL.Path == fmt.Sprintf("/api/core/v1/users/%s/globalroles", user) {
		fc.ProcessFilter(req, resp)
		return
	}

	runtimeCache := cache.GetCurrentCache()
	tmpl, err := runtimeCache.GetTemplate()
	if err != nil {
		apiUtil.ResponseError(resp, http.StatusInternalServerError, err.Error())
		return
	} else if tmpl == nil {
		apiUtil.ResponseError(resp, http.StatusBadRequest, "No template found")
		return
	} else if tmpl.KSHost == "" {
		apiUtil.ResponseError(resp, http.StatusBadRequest, "caas auth not init")
		return
	}

	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}
	// TODO: scheme is https ?
	caasUrl := &url.URL{Scheme: "http", Host: tmpl.KSHost, Path: fmt.Sprintf("kapis/iam.kubesphere.io/v1alpha2/users/%s/globalroles", user)}

	httpresp, code, err := util.CommonRequest(caasUrl.String(), http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		apiUtil.ResponseError(resp, http.StatusInternalServerError, err.Error())
		return
	}
	switch code {
	case http.StatusUnauthorized:
		apiUtil.ResponseError(resp, http.StatusUnauthorized, fmt.Sprintf("user %s unauthorized", user))
		return
	case http.StatusForbidden:
		apiUtil.ResponseError(resp, http.StatusUnauthorized, fmt.Sprintf("user %s forbidden", user))
		return
		// case http.StatusOK:
		// 	fallthrough
		// default:
		// 	//todo..
	}

	//TODO: buffer cache for performance
	caaSGlobalRoles := make([]schema.CaaSGlobalRoles, 0)
	err = json.Unmarshal(httpresp, &caaSGlobalRoles)
	if err != nil {
		apiUtil.ResponseError(resp, http.StatusBadRequest, err.Error())
		return
	}
	allowed := false
	for _, globalrole := range caaSGlobalRoles {
		for _, rule := range globalrole.Rules {
			if APIGroupMatches(&rule, "core") &&
				verbMatches(&rule, req.Request.Method) &&
				ResourcesMatches(&rule, "*") {
				allowed = true
				break
			}
		}
		if allowed {
			break
		}
	}

	if allowed {
		fc.ProcessFilter(req, resp)
		return
	}
	apiUtil.ResponseError(resp, http.StatusForbidden, fmt.Sprintf("user %s forbidden", user))
}

func getToken(req *restful.Request) (string, error) {
	headerToken := req.HeaderParameter("token")
	if headerToken == "" {
		cookie, err := req.Request.Cookie("token")
		if err != nil {
			return headerToken, err
		} else {
			headerToken = cookie.Value
		}
	}
	return headerToken, nil
}

func getUserHeader(req *restful.Request) (string, error) {
	user := req.HeaderParameter(constants.UserHeader)
	if user == "" {
		return "", fmt.Errorf("no user found in header with caas auth type")
	}
	return user, nil
}

func verbMatches(rule *schema.CaaSRules, requestedVerb string) bool {
	for _, ruleVerb := range rule.Verbs {
		if ruleVerb == "*" {
			return true
		}
		if ruleVerb == requestedVerb {
			return true
		}
	}
	return false
}

func APIGroupMatches(rule *schema.CaaSRules, requestedGroup string) bool {
	for _, ruleGroup := range rule.APIGroups {
		if ruleGroup == "*" {
			return true
		}
		if ruleGroup == requestedGroup {
			return true
		}
	}

	return false
}

// TODO: now always is *
func ResourcesMatches(rule *schema.CaaSRules, requestedResource string) bool {
	for _, resource := range rule.Resources {
		if resource == "*" {
			return true
		}
		if resource == requestedResource {
			return true
		}
	}

	return false
}
