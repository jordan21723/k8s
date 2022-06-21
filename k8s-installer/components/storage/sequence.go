package storage

import (
	"k8s-installer/schema"
)

func SetDefaultStorageClass(cluster *schema.Cluster) {
	if cluster.CloudProvider != nil {
		if cluster.CloudProvider.OpenStack != nil {
			if cluster.CloudProvider.OpenStack.Enable {
				cluster.CloudProvider.OpenStack.Cinder.IsDefaultSc = true
				return
			}
		}
	}

	if cluster.Storage.NFS != nil {
		if cluster.Storage.NFS.Enable {
			cluster.Storage.NFS.IsDefaultSc = true
		}
	}

	if cluster.Storage.Ceph != nil {
		if cluster.Storage.Ceph.Enable {
			cluster.Storage.Ceph.IsDefaultSc = true
		}
	}
}
