package utils

import (
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
	"net/http"

	"github.com/emicklei/go-restful"
)

func SchemaValidationFailedResponseError(resp *restful.Response, errorMessage string) {
	log.Errorf("Schema validation failed: %s", errorMessage)
	_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, &schema.HttpErrorResult{
		ErrorCode: http.StatusBadRequest,
		Message:   errorMessage,
	})
}

func ResponseError(resp *restful.Response, httpStatusCode int, errorMessage string) {
	log.Errorf("Got api call error: %s", errorMessage)
	_ = resp.WriteHeaderAndEntity(httpStatusCode, &schema.HttpErrorResult{
		ErrorCode: httpStatusCode,
		Message:   errorMessage,
	})
}

func ResponseErrorWithCode(resp *restful.Response, httpStatusCode int, errorCode int, errorMessage string) {
	log.Errorf("Got api call error: %s", errorMessage)
	_ = resp.WriteHeaderAndEntity(httpStatusCode, &schema.HttpErrorResult{
		ErrorCode: errorCode,
		Message:   errorMessage,
	})
}

func ResponseErrorWithList(resp *restful.Response, errorCode int, errorMessage string, errorList []string) {
	log.Errorf("Got api call error: %s", errorMessage)
	_ = resp.WriteHeaderAndEntity(errorCode, &schema.HttpErrorResult{
		ErrorCode: errorCode,
		Message:   errorMessage,
		ErrorList: errorList,
	})
}

func WriteSingleEmptyStructWithStatus(resp *restful.Response) {
	_, _ = resp.Write([]byte("{}"))
}
