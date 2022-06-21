package template

import (
	"bytes"
	"text/template"

	"k8s-installer/pkg/util"
)

type DeployTemplate struct {
	template.Template
}

const (
	Base64EncodePipelineKey = "b64enc"
	SplitPipelineKey        = "split"
	StrArrIndexPipelineKey  = "index"
)

var funcMap = template.FuncMap{
	Base64EncodePipelineKey: util.Base64Encode,
	SplitPipelineKey:        util.StringSplit,
	StrArrIndexPipelineKey:  util.StringArrayLocate,
}

func New(name string) *DeployTemplate {
	return &DeployTemplate{
		Template: *template.New(name).Funcs(funcMap),
	}
}

func (dt *DeployTemplate) Render(tempStr string, data interface{}) (string, error) {
	temp, err := dt.Parse(tempStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = temp.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
