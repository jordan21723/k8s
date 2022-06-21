package user

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/util"

	"k8s-installer/pkg/cache"
	jwtPkg "k8s-installer/pkg/jwt"
	"k8s-installer/schema"

	"github.com/dgrijalva/jwt-go"
	"github.com/emicklei/go-restful"
)

func SendMesssageCode(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	conf := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	if conf.Mock {
		response.WriteAsJson(&schema.SendMesssageResp{Code: 0, DataObject: schema.SendDataObject{RandomCode: ""}})
		return
	}

	messsageType := request.QueryParameter("type")
	if messsageType != "102" && messsageType != "103" {
		utils.ResponseError(response, http.StatusBadRequest, "query parameter 'type' missing")
		return
	}
	userPost := schema.User{}
	if err := request.ReadEntity(&userPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	phone := userPost.Phone
	if phone == "" {
		userInfo, err := runtimeCache.GetUser(userPost.Username)
		if err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information due to error: %s", err.Error()))
			return
		}
		if userInfo == nil {
			utils.ResponseError(response, http.StatusInternalServerError, "User does not exist")
			return
		}
		if userInfo.Phone == "" {
			utils.ResponseError(response, http.StatusInternalServerError, "User Phone is not set")
			return
		}

		phone = userInfo.Phone
	}

	msgUrl := url.URL{
		Scheme:   "http",
		Host:     conf.MesssagePlatformHost,
		Path:     "tymh_interface_sms/random_code_new",
		RawQuery: fmt.Sprintf("id=%s&type=%s", phone, messsageType),
	}
	log.Debug("msgUrl: ", msgUrl.String())
	msgResp, err := SendMsgHttpGet(msgUrl.String())
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Messsage request failed: %s", err.Error()))
		return
	}
	response.WriteAsJson(msgResp)
}

func ValidateMesssageCode(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	conf := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	if conf.Mock {
		response.WriteAsJson(&schema.ValidateMesssageResp{Code: 0, DataObject: schema.ValidateDataObject{Result: true}})
		return
	}

	msgPost := schema.ValidateMesssageCode{}
	if err := request.ReadEntity(&msgPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	phone := msgPost.Phone
	if phone == "" {
		userInfo, err := runtimeCache.GetUser(msgPost.UserName)
		if err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information due to error: %s", err.Error()))
			return
		}
		if userInfo == nil {
			utils.ResponseError(response, http.StatusUnauthorized, "Username or password incorrect")
			return
		}
		if userInfo.Phone == "" {
			utils.ResponseError(response, http.StatusUnauthorized, "User Phone is not set")
			return
		}
		phone = userInfo.Phone
	}

	msgUrl := url.URL{
		Scheme:   "http",
		Host:     conf.MesssagePlatformHost,
		Path:     "tymh_interface_sms/random_code/verify",
		RawQuery: fmt.Sprintf("id=%s&type=%s&randomcode=%s", phone, msgPost.Type, msgPost.RandomCode),
	}
	msgResp, err := ValidateMsgHttpGet(msgUrl.String())
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Messsage request failed: %s", err.Error()))
		return
	}
	response.WriteAsJson(msgResp)
}

func SendMsgHttpGet(url string) (*schema.SendMesssageResp, error) {
	resp, err := util.HttpGet(url, time.Second*3)
	log.Debug("HttpGet error: ", err)
	if err != nil {
		return nil, err
	}
	byteData, err := ioutil.ReadAll(resp.Body)
	log.Debug("ReadAll error: ", err)
	log.Debug("byteData : ", string(byteData))
	if err != nil {
		return nil, err
	}
	msgResp := schema.SendMesssageResp{}
	err = json.Unmarshal(byteData, &msgResp)
	log.Debug("Unmarshal error: ", err)
	return &msgResp, err
}

func ValidateMsgHttpGet(url string) (*schema.ValidateMesssageResp, error) {
	resp, err := util.HttpGet(url, time.Second*3)
	log.Debug("HttpGet error: ", err)
	if err != nil {
		return nil, err
	}
	byteData, err := ioutil.ReadAll(resp.Body)
	log.Debug("ReadAll error: ", err)
	log.Debug("byteData : ", string(byteData))
	if err != nil {
		return nil, err
	}
	msgResp := schema.ValidateMesssageResp{}
	err = json.Unmarshal(byteData, &msgResp)
	log.Debug("Unmarshal error: ", err)
	return &msgResp, err
}

func GetAuthType(request *restful.Request, response *restful.Response) {
	ksHost, err := getKsHost()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	}
	resp := schema.SystemAuthType{
		Standard: true,
		CaaS:     ksHost != "",
	}
	response.WriteHeaderAndEntity(http.StatusOK, resp)
}

func standardLogin(request *restful.Request, response *restful.Response) {
	userPost := schema.User{}
	if err := request.ReadEntity(&userPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	if userPost.Password == "" {
		utils.ResponseError(response, http.StatusBadRequest, "Password cannot be empty")
		return
	}
	runtimeCache := cache.GetCurrentCache()
	user, err := runtimeCache.Login(userPost.Username, userPost.Password)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information due to error: %s", err.Error()))
		return
	}

	if user == nil {
		utils.ResponseError(response, http.StatusUnauthorized, "Username or password incorrect")
		return
	}

	rolesMapping, errRole := runtimeCache.GetRoleList()
	if errRole != nil {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to get role list from db during login")
		log.Errorf("Failed to get role list from db during authentication check due to error: %s", errRole.Error())
		return
	}

	var userPermission uint64
	for index, role := range user.Roles {
		if roleFound, isFound := rolesMapping[role.Id]; isFound {
			userPermission |= roleFound.Function
			user.Roles[index] = rolesMapping[role.Id]
		}
	}

	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	signString, err := base64.StdEncoding.DecodeString(config.ApiServer.JWTSignString)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to decode jwt sign string")
		return
	}

	expired := time.Now().Add(time.Duration(config.ApiServer.JWTExpiredAfterHours) * time.Hour).Unix()
	user.StandardClaims = jwt.StandardClaims{
		ExpiresAt: expired,
	}
	signedToken, errJWT := jwtPkg.CreateJWT(user, string(signString), config.ApiServer.JWTSignMethod)
	if errJWT != nil {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to create jwt token")
		return
	}

	// unset password for security consideration
	user.Password = ""

	expiredString := time.Unix(expired, 0).Format("2006-01-02T15:04:05Z07:00")
	response.WriteAsJson(struct {
		Token       string      `json:"token"`
		ExpiredDate string      `json:"expired_date"`
		Permission  uint64      `json:"permission"`
		User        schema.User `json:"user"`
	}{
		Token:       signedToken,
		ExpiredDate: expiredString,
		Permission:  userPermission,
		User:        *user,
	})
}

func Login(request *restful.Request, response *restful.Response) {
	switch request.HeaderParameter(constants.AuthTypeHeader) {
	case constants.StandardAuth:
		standardLogin(request, response)
	case constants.CaaSAuth:
		ksOauth(request, response)
	default:
		// wait front-end integrate, then return error , now workarround
		// utils.ResponseError(response, http.StatusInternalServerError, "This authentication method does not support")
		standardLogin(request, response)
	}
}

func ksOauth(request *restful.Request, response *restful.Response) {
	user := schema.User{}
	if err := request.ReadEntity(&user); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	ksHost, err := getKsHost()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, err.Error())
		return
	} else if ksHost == "" {
		utils.ResponseError(response, http.StatusBadRequest, "caas auth not init")
		return
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/x-www-form-urlencoded"

	dataUrlVal := url.Values{}
	dataUrlVal.Add("grant_type", "password")
	dataUrlVal.Add("username", user.Username)
	dataUrlVal.Add("password", user.Password)

	ksUrl := &url.URL{Scheme: "http", Host: ksHost, Path: "oauth/token"}

	ksUser := &schema.CaaSOauthRequest{}

	dataByte := bytes.NewBufferString(dataUrlVal.Encode()).Bytes()

	resp, code, err := util.CommonRequest(ksUrl.String(), http.MethodPost, "", dataByte, header, true, true, 0)
	if err != nil {
		utils.ResponseError(response, code, err.Error())
		return
	}
	switch code {
	case http.StatusUnauthorized:
		utils.ResponseError(response, code, "unauthorized")
		return
	case 200:
		fallthrough
	default:
		log.Debug("caas auth resp code is ", code)
		// TODO: other code
	}
	if err = json.Unmarshal(resp, ksUser); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteAsJson(ksUser)
}

func CreateOrUpdateUser(request *restful.Request, response *restful.Response) {
	userPost := schema.User{}
	if err := request.ReadEntity(&userPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	userPost.Username = strings.TrimSpace(userPost.Username)
	if request.Request.Method == "POST" {
		userPost.Id = fmt.Sprintf("%v", hash(userPost.Username))
	} else {
		username := strings.TrimSpace(request.PathParameter("username"))
		if strings.TrimSpace(username) == "" {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "user-id"))
			return
		}
		userPost.Username = username
	}

	err := utils.Validate(userPost)
	if err != nil {
		utils.SchemaValidationFailedResponseError(response, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	u, err := runtimeCache.GetUser(userPost.Username)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user due to %v", err))
		return
	}

	if u != nil {

		if request.Request.Method == "PUT" {
			// ensure admin cannot be change
			if u.Username == strings.ToLower("admin") {
				utils.ResponseError(response, http.StatusBadRequest, "Cannot create or modify internal user admin")
				return
			}
			// check role id exists
			roleCheckResult := checkRoleExistence(userPost.Roles)
			u.Roles = userPost.Roles
			u.Password = userPost.Password
			u.Phone = userPost.Phone
			u.KsAccount = u.KsAccount

			if len(roleCheckResult) > 0 {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("(ignored) Following roles id check failed: %s", strings.Join(roleCheckResult, " , ")))
				//return
			}
			if err = runtimeCache.CreateOrUpdateUser(userPost.Username, *u); err != nil {
				utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create new user due to %v", err))
				return
			} else {
				response.WriteAsJson(u)
				return
			}
		} else {
			if userPost.Username == strings.ToLower("admin") {
				utils.ResponseError(response, http.StatusBadRequest, "Cannot create or modify internal user admin")
				return
			} else {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("User %v already exists", u.Username))
				return
			}

		}
	} else {
		if request.Request.Method == http.MethodPut {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find user %s", userPost.Username))
			return
		} else {
			// ensure admin cannot be changed
			if userPost.Username == strings.ToLower("admin") {
				utils.ResponseError(response, http.StatusBadRequest, "Cannot create or modify internal user admin")
				return
			}

			// check role id exists
			roleCheckResult := checkRoleExistence(userPost.Roles)
			if len(roleCheckResult) > 0 {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Following roles id check failed: %s", strings.Join(roleCheckResult, " , ")))
				return
			}
			if err = runtimeCache.CreateOrUpdateUser(userPost.Id, userPost); err != nil {
				utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create new user due to %v", err))
				return
			} else {
				response.WriteAsJson(userPost)
			}
		}

	}
}

func ListUser(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	ul, err := runtimeCache.GetUserList()
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to get user list due to %v", err),
		})
	}

	rolesMapping, errRole := runtimeCache.GetRoleList()
	if errRole != nil {
		utils.ResponseError(response, http.StatusInternalServerError, "Failed to get role list from db during listing user")
		return
	}

	var ulj []schema.User
	for _, u := range ul {
		// load role data from db with id
		for index, role := range u.Roles {
			if roleFound, isFound := rolesMapping[role.Id]; isFound {
				u.Roles[index].Name = roleFound.Name
				u.Roles[index].Function = roleFound.Function
			}
		}

		ulj = append(ulj, u)
	}

	response.WriteAsJson(ulj)
}

func DeleteUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	if strings.TrimSpace(username) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "username"))
		return
	}
	runtimeCache := cache.GetCurrentCache()

	userFound, err := runtimeCache.GetUser(username)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information with username: %s due to error: %s", username, err.Error()))
		return
	}

	if userFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found username: %s", username))
		return
	}

	if strings.ToLower(userFound.Username) == "admin" {
		utils.ResponseError(response, http.StatusBadRequest, "Cannot delete internal user admin")
		return
	}

	runtimeCache.DeleteUser(username)
	response.WriteHeader(http.StatusOK)
}

func getStandardUser(request *restful.Request, response *restful.Response) {
	username := request.PathParameter("username")
	if strings.TrimSpace(username) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "username"))
		return
	}
	runtimeCache := cache.GetCurrentCache()
	user, err := runtimeCache.GetUser(username)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get user information due to error: %s", err.Error()))
		return
	}
	if user == nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("User with id %v does not exist.", username))
		return
	}

	response.WriteAsJson(user)
}

func GetUser(request *restful.Request, response *restful.Response) {
	switch request.HeaderParameter(constants.AuthTypeHeader) {
	case constants.StandardAuth:
		getStandardUser(request, response)
	case constants.CaaSAuth:
		getCaaSUser(request, response)
	default:
		// wait front-end integrate, then return error , now workarround
		// utils.ResponseError(response, http.StatusInternalServerError, "This authentication method does not support")
		getStandardUser(request, response)
	}
}

func getCaaSUser(request *restful.Request, response *restful.Response) {
	resp, code, err := ksQueryUrl(
		fmt.Sprintf("kapis/iam.kubesphere.io/v1alpha2/users/%s",
			request.PathParameter("username")),
		request.HeaderParameter("token"))
	if err != nil {
		utils.ResponseError(response, code, err.Error())
		return
	}

	ksUser := new(schema.CaaSUser)
	err = json.Unmarshal(resp, ksUser)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	response.WriteAsJson(ksUser)
}

func GetUserGlobalRules(request *restful.Request, response *restful.Response) {
	resp, code, err := ksQueryUrl(
		fmt.Sprintf("kapis/iam.kubesphere.io/v1alpha2/users/%s/globalroles",
			request.PathParameter("username")),
		request.HeaderParameter("token"))
	if err != nil {
		utils.ResponseError(response, code, err.Error())
		return
	}

	ksRules := make([]schema.CaaSGlobalRoles, 0)
	err = json.Unmarshal(resp, &ksRules)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	sig := true
	kr := make([]schema.CaaSGlobalRoles, 0)
	for _, data := range ksRules {
		for _, rule := range data.Rules {
			if sig {
				for _, group := range rule.APIGroups {
					if group == "core" {
						kr = append(kr, data)
						sig = !sig
						break
					}
				}
			} else {
				break
			}
		}
		sig = !sig
	}

	response.WriteAsJson(kr)
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func CreateOrUpdateRole(request *restful.Request, response *restful.Response) {
	rolePost := schema.Role{}
	if err := request.ReadEntity(&rolePost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}
	if rolePost.Function == 0 {
		utils.ResponseError(response, http.StatusBadRequest, "Create a none function role is not allowed")
		return
	}

	if rolePost.Function > 134217727 {
		utils.ResponseError(response, http.StatusBadRequest, "Function code greater than 134217727 is not allowed")
		return
	}

	rolePost.Name = strings.TrimSpace(rolePost.Name)
	if request.Request.Method == "POST" {
		rolePost.Id = fmt.Sprintf("%v", hash(rolePost.Name))
	} else {
		roleId := request.PathParameter("role-id")
		if strings.TrimSpace(roleId) == "" {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "role-id"))
			return
		}
		rolePost.Id = roleId
	}

	err := utils.Validate(rolePost)
	if err != nil {
		utils.SchemaValidationFailedResponseError(response, err.Error())
		return
	}

	v := reflect.ValueOf(rolePost)
	count := v.NumField()
	for i := 0; i < count; i++ {
		f := v.Field(i)
		if f.String() == "" {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("%v cannot be empty", f.Elem().Type().Name()))
			return
		}
	}
	runtimeCache := cache.GetCurrentCache()
	r, err := runtimeCache.GetRole(rolePost.Id)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get role due to %v", err))
		return
	}

	if r != nil {
		if request.Request.Method == "PUT" {
			if strings.ToLower(r.Name) == "admin" {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot modify internal role admin"))
				return
			}
			r.Function = rolePost.Function
			//Role name is not allowed to change
			rolePost.Name = r.Name
			if err = runtimeCache.CreateOrUpdateRole(rolePost.Id, rolePost); err != nil {
				utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create new role due to %v", err))
				return
			}
		} else {
			if strings.ToLower(rolePost.Name) == "admin" {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot create internal role admin"))
				return
			}
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Role %v already exists: %v", r.Name, err))
			return
		}
	} else {
		if request.Request.Method == "PUT" {
			utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found role with id: %s", rolePost.Id))
			return
		} else {
			if strings.ToLower(rolePost.Name) == "admin" {
				utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot create internal role admin"))
				return
			}
			if err = runtimeCache.CreateOrUpdateRole(rolePost.Id, rolePost); err != nil {
				utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create new role due to %v", err))
				return
			}
		}

	}
	response.WriteAsJson(rolePost)
}

func ListRole(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	rl, err := runtimeCache.GetRoleList()
	if err != nil {
		_ = response.WriteHeaderAndEntity(http.StatusInternalServerError, &schema.HttpErrorResult{
			ErrorCode: http.StatusInternalServerError,
			Message:   fmt.Sprintf("Failed to get role list due to %v", err),
		})
	}
	var rlj []schema.Role
	for _, r := range rl {
		rlj = append(rlj, r)
	}

	response.WriteAsJson(rlj)
}

func DeleteRole(request *restful.Request, response *restful.Response) {
	roleId := request.PathParameter("role-id")
	if strings.TrimSpace(roleId) == "" {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find path parameter: %s", "role-id"))
		return
	}
	runtimeCache := cache.GetCurrentCache()

	roleFound, err := runtimeCache.GetRole(roleId)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found role with id: %s", roleId))
		return
	}

	if roleFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found role with id: %s", roleId))
		return
	}

	if strings.ToLower(roleFound.Name) == "admin" {
		utils.ResponseError(response, http.StatusBadRequest, "Cannot delete internal role admin")
		return
	}

	runtimeCache.DeleteRole(roleId)
}

func GetRole(request *restful.Request, response *restful.Response) {
	roleId := request.PathParameter("role-id")
	if strings.TrimSpace(roleId) == "" {
		_ = response.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
			ErrorCode: http.StatusBadRequest,
			Message:   fmt.Sprintf("Unable to find path parameter: %s", "role-id"),
		})
		return
	}
	runtimeCache := cache.GetCurrentCache()

	roleFound, err := runtimeCache.GetRole(roleId)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found role with id: %s", roleId))
		return
	}

	if roleFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot found role with id: %s", roleId))
		return
	}

	response.WriteAsJson(roleFound)
}

func checkRoleExistence(roles []schema.Role) []string {
	var result []string
	runtimeCache := cache.GetCurrentCache()
	for _, role := range roles {
		roleFound, err := runtimeCache.GetRole(role.Id)
		if err != nil {
			result = append(result, err.Error())
		} else if roleFound == nil {
			result = append(result, fmt.Sprintf("Unable found role with id %s", role.Id))
		}
	}
	return result
}

func getKsHost() (string, error) {
	runtimeCache := cache.GetCurrentCache()
	tmpl, err := runtimeCache.GetTemplate()
	if err != nil {
		return "", fmt.Errorf("Failed to get templete information due to error: %s", err.Error())
	} else if tmpl == nil {
		return "", errors.New("No template found")
	}
	return tmpl.KSHost, nil
}

func ksQueryUrl(path, ksToken string) ([]byte, int, error) {
	host, err := getKsHost()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	} else if host == "" {
		return nil, http.StatusInternalServerError, errors.New("caas auth not init")
	}
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", ksToken),
	}
	ksUrl := &url.URL{Scheme: "http", Host: host, Path: path}

	resp, code, err := util.CommonRequest(ksUrl.String(), http.MethodGet, "", nil, header, true, true, 0)
	if err != nil {
		return resp, code, err
	}
	if code != http.StatusOK {
		err = errors.New(string(resp))
	}
	return resp, code, err
}
