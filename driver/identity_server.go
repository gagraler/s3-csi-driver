package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"
	pb "google.golang.org/protobuf/types/known/wrapperspb"
	common "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 14:42:11
 * @file: identity_server.go
 * @description: 身份服务
 */

type IdentityServer struct {
	*common.DefaultIdentityServer
}

// Probe detect whether the plugin is running.
// 探测检查插件是否正在运行, 此方法不需要返回任何内容, 而且截至目前，规范也没有规定该返回什么, so, 返回nil
func (i *IdentityServer) Probe(_ context.Context, _ *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: &pb.BoolValue{
			Value: true,
		},
	}, nil
}

// GetPluginCapabilities returns the capabilities of the plugin.
func (i *IdentityServer) GetPluginCapabilities(_ context.Context, _ *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
