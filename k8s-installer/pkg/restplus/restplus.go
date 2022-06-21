package restplus

import (
	"strconv"

	"github.com/emicklei/go-restful"
)

func GetBoolValueWithDefault(req *restful.Request, name string, dv bool) bool {
	reverse := req.QueryParameter(name)
	if v, err := strconv.ParseBool(reverse); err == nil {
		return v
	}
	return dv
}

func GetStringValueWithDefault(req *restful.Request, name string, dv string) string {
	v := req.QueryParameter(name)
	if v == "" {
		v = dv
	}
	return v
}

func GetWrapperBoolWithDefault(req *restful.Request, name string, dv *bool) *bool {
	reverse := req.QueryParameter(name)
	if v, err := strconv.ParseBool(reverse); err == nil {
		return &v
	}
	return dv
}

func GetIntValueWithDefault(req *restful.Request, name string, dv int) int {
	v := req.QueryParameter(name)
	if v, err := strconv.ParseInt(v, 10, 64); err == nil {
		return int(v)
	}
	return dv
}
