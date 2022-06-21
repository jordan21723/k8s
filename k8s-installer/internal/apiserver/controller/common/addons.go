package common

import (
	"k8s-installer/schema"
	"k8s-installer/schema/plugable"
)

func SetDependencies(cluster *schema.Cluster) {
	if cluster.Console != nil {
		cluster.Console.SetDependencies(map[string]plugable.IPlugAble{
			"middle-platform": cluster.MiddlePlatform,
			"efk":             cluster.EFK,
			"gap":             cluster.GAP,
		})
	}

	if cluster.MiddlePlatform != nil {
		cluster.MiddlePlatform.SetDependencies(map[string]plugable.IPlugAble{
			"pgo": cluster.PostgresOperator,
		})
	}

	if cluster.GAP != nil {
		cluster.GAP.SetDependencies(map[string]plugable.IPlugAble{
			"helm": cluster.Helm,
		})
	}
}
