package cache

import (
	"k8s-installer/schema"
)

// config  no need to lock down
// it does not required db operation
// it`s feel to access
var serverRuntimeConfigCollection = schema.ServerRuntimeConfigCollection{}
var clientRuntimeConfigCollection = schema.ClientRuntimeConfigCollection{}
var NodeId string
