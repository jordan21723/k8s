package nfs

import (
	"k8s-installer/schema/plugable"
	"testing"
)

func TestNFS_StorageClassTemplateRender(t *testing.T) {
	type fields struct {
		Enable           bool
		NFSServerAddress string
		NFSPath          string
		ReclaimPolicy    string
		ImageRegistry    string
		StorageClassName string
		Dependencies     map[string]plugable.IPlugAble
		Status           string
		IsDefaultSc      bool
		MountOptions     []string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "templete MountOptions test",
			fields: fields{
				Enable:           true,
				NFSServerAddress: "127.0.0.1",
				NFSPath:          "/tmp/nfs",
				ReclaimPolicy:    "Delete",
				ImageRegistry:    "127.0.0.1:5000",
				StorageClassName: "nfs",
				Dependencies:     map[string]plugable.IPlugAble{},
				Status:           "auto",
				IsDefaultSc:      true,
				MountOptions:     []string{"vers=3", "other"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfs := &NFS{
				Enable:           tt.fields.Enable,
				NFSServerAddress: tt.fields.NFSServerAddress,
				NFSPath:          tt.fields.NFSPath,
				ReclaimPolicy:    tt.fields.ReclaimPolicy,
				ImageRegistry:    tt.fields.ImageRegistry,
				StorageClassName: tt.fields.StorageClassName,
				Dependencies:     tt.fields.Dependencies,
				Status:           tt.fields.Status,
				IsDefaultSc:      tt.fields.IsDefaultSc,
				MountOptions:     tt.fields.MountOptions,
			}
			got, err := nfs.StorageClassTemplateRender()
			if err != nil {
				t.Errorf("StorageClassTemplateRender() error = %v", err)
				return
			}
			t.Log("templete: ", got)
		})
	}
}
