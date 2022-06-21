package schema

import (
	"k8s-installer/components/storage/ceph"
	"k8s-installer/components/storage/nfs"
)

type Storage struct {
	NFS  *nfs.NFS   `json:"nfs" description:"nfs section"`
	Ceph *ceph.Ceph `json:"ceph" description:"ceph section"`
}
