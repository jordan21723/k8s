package schema

import "github.com/dgrijalva/jwt-go"

type HttpErrorResult struct {
	// error code is custom error definition code to help caller understand error
	ErrorCode int      `json:"error_code"`
	Message   string   `json:"error_message"`
	ErrorList []string `json:"error_list"`
}

type User struct {
	Id        string `json:"id,omitempty" description:"do not input auto generator"`
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	Roles     []Role `json:"roles,omitempty" description:"contain one or more roles"`
	Phone     string `json:"phone" description:"user phone number"`
	KsAccount string `json:"ks_account,omitempty" description:"kubespere user account"`
	jwt.StandardClaims
}

type Role struct {
	Id       string `json:"id,omitempty" description:"do not input auto generator"`
	Name     string `json:"name" validate:"required"`
	Function uint64 `json:"function" validate:"required" description:"function code represent capability of the role"`
}

type CaaSOauthRequest struct {
	TokenType    string `json:"token_type,omitempty" description:"Token Type: Bearer"`
	AccessToken  string `json:"access_token,omitempty" description:" ks api token"`
	RefreshToken string `json:"refresh_token,omitempty" description:" ks refresh token"`
}

type CaaSUser struct {
	MetaData CaaSUserMetadata `json:"metadata,omitempty" description:"user detail metadata"`
	Spec     CaaSUserSpec     `json:"spec,omitempty" description:"user detail spec"`
}

type CaaSUserSpec struct {
	Email string `json:"email,omitempty" description:"user detail email"`
}

type CaaSUserMetadata struct {
	UserName    string                      `json:"name,omitempty" description:"user detail user name"`
	Annotations CaaSUserMetadataAnnotations `json:"annotations,omitempty" description:"user detail annotations"`
}

type CaaSUserMetadataAnnotations struct {
	GlobalRole string `json:"iam.kubesphere.io/globalrole,omitempty" description:"user detail annotations GlobalRole list"`
}

type CaaSGlobalRoles struct {
	MetaData CaasRuleMetadata `json:"metadata,omitempty" description:"user rule metadata"`
	Rules    []CaaSRules      `json:"rules,omitempty" description:"user rule list"`
}

type CaasRuleMetadata struct {
	Name        string                      `json:"name,omitempty" description:"user rule metadata-name"`
	Labels      CaaSRuleMetadataLabels      `json:"labels,omitempty" description:"user rule metadata-labels"`
	Annotations CaaSRuleMetadataAnnotations `json:"annotations,omitempty" description:"user rule-Annotations"`
}

type CaaSRuleMetadataLabels struct {
	RoleTemplate string `json:"iam.kubesphere.io/role-template,omitempty" description:"user rule-RoleTemplate"`
	Managed      string `json:"kubefed.io/managed,omitempty" description:"user rule-Managed"`
}

type CaaSRuleMetadataAnnotations struct {
	Dependencies      string `json:"iam.kubesphere.io/dependencies,omitempty" description:"user rule-Annotations Dependencies"`
	Module            string `json:"iam.kubesphere.io/module,omitempty" description:"user rule-Annotations Module"`
	RoleTemplateRules string `json:"iam.kubesphere.io/role-template-rules,omitempty" description:"user rule-Annotations RoleTemplateRules"`
}

type CaaSRules struct {
	Verbs     []string `json:"verbs,omitempty" description:"user rule-Verbs"`
	APIGroups []string `json:"apiGroups,omitempty" description:"user rule-APIGroups"`
	Resources []string `json:"resources,omitempty" description:"user rule-Resources"`
}

type SystemAuthType struct {
	Standard bool `json:"standard" description:"local auth"`
	CaaS     bool `json:"caas" description:"caas redirect auth"`
}

type SendMesssageResp struct {
	Code             int            `json:"code"`
	ErrorDescription string         `json:"errorDescription"`
	DataObject       SendDataObject `json:"dataObject"`
}

type SendDataObject struct {
	RandomCode string `json:"randomCode"`
}

type ValidateMesssageResp struct {
	Code             int                `json:"code"`
	ErrorDescription string             `json:"errorDescription"`
	DataObject       ValidateDataObject `json:"dataObject"`
}

type ValidateDataObject struct {
	Result bool `json:"result"`
}

type ValidateMesssageCode struct {
	UserName   string `json:"username" validate:"required"`
	RandomCode string `json:"randomcode" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Phone      string `json:"phone"`
}
