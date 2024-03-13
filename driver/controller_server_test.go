package driver

import (
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-12 17:04:17
 * @file: controller_server_test.go
 * @description: controller_server 单测
 */

func TestNewControllerServiceCapability(t *testing.T) {
	tests := []struct {
		name string
		cap  csi.ControllerServiceCapability_RPC_Type
		want *csi.ControllerServiceCapability
	}{
		{
			name: "Test CREATE_DELETE_VOLUME",
			cap:  csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			want: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
		},
		{
			name: "Test PUBLISH_UNPUBLISH_VOLUME",
			cap:  csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
			want: &csi.ControllerServiceCapability{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewControllerServiceCapability(tt.cap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewControllerServiceCapability() = %v, want %v", got, tt.want)
			}
		})
	}
}
