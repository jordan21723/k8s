package plugable

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func CommonCheckDependency(plugable IPlugAble) (error,[]string) {
	var errorMsgs []string
	if !plugable.IsEnable() {
		return nil, nil
	}
	for name, dep := range plugable.GetDependencies() {
		if dep == nil || reflect.ValueOf(dep).IsNil() {
			errorMsgs = append(errorMsgs, fmt.Sprintf("Plugin %s require %s", plugable.GetName(), name))
			continue
		} else if dep.IsEnable() == false {
			errorMsgs = append(errorMsgs, fmt.Sprintf("Plugin %s require %s", plugable.GetName(), name))
			continue
		}
	}
	if len(errorMsgs) > 0 {
		return errors.New(fmt.Sprintf("Dependency check failed due to error %s", strings.Join(errorMsgs, "\n"))), errorMsgs
	} else {
		return nil, nil
	}
}