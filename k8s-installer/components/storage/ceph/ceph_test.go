package ceph

import (
	"k8s-installer/schema/plugable"
	"testing"
)

func TestCeph_DeploymentTemplateRender(t *testing.T) {
	type fields struct {
		Enable           bool
		CephClusterId    string
		MonitorIPList    []string
		PoolUserId       string
		PoolUserKey      string
		StorageClassName string
		ReclaimPolicy    string
		ImageRegistry    string
		Dependencies     map[string]plugable.IPlugAble
		Status           string
		IsDefaultSc      bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "test base",
			fields: fields{
				Enable:           true,
				CephClusterId:    "a0dce530-f53e-11eb-853d-fa163ea5e697",
				MonitorIPList:    []string{"172.20.151.105:6789", "172.20.150.200:6789", "172.20.149.204:6789"},
				PoolUserId:       "kubernetes",
				PoolUserKey:      "AQAs7AxhUwDBHBAAzrkWTdsLpJdqF5k0YZPgEg==",
				StorageClassName: "ceph",
				ReclaimPolicy:    "delete",
				ImageRegistry:    "172.20.149.95:5000",
				IsDefaultSc:      true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Ceph{
				Enable:           tt.fields.Enable,
				CephClusterId:    tt.fields.CephClusterId,
				MonitorIPList:    tt.fields.MonitorIPList,
				PoolUserId:       tt.fields.PoolUserId,
				PoolUserKey:      tt.fields.PoolUserKey,
				StorageClassName: tt.fields.StorageClassName,
				ReclaimPolicy:    tt.fields.ReclaimPolicy,
				ImageRegistry:    tt.fields.ImageRegistry,
				Dependencies:     tt.fields.Dependencies,
				Status:           tt.fields.Status,
				IsDefaultSc:      tt.fields.IsDefaultSc,
			}
			got, err := c.DeploymentTemplateRender()
			if err != nil {
				t.Errorf("DeploymentTemplateRender() error = %v", err)
				return
			}
			t.Log("got: ", got)
		})
	}
}
