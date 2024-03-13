package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/keington/s3-csi-driver/driver/utils"
	mounter "github.com/keington/s3-csi-driver/driver/utils/mounter"

	"github.com/container-storage-interface/spec/lib/go/csi"
	common "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 16:48:46
 * @file: node_server.go
 * @description: 节点服务
 */

type NodeServer struct {
	*common.DefaultNodeServer
}

// NodeGetInfo implements csi.NodeServer.
// Returns the supported capabilities of the node server.
func (n *NodeServer) NodeGetInfo(_ context.Context, _ *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetInfo is not implemented")
}

// NodeGetCapabilities implements csi.NodeServer.
// Returns the supported capabilities of the node server.
func (n *NodeServer) NodeGetCapabilities(_ context.Context, _ *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	nodeServerCapabilities := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{nodeServerCapabilities},
	}, nil
}

// NodeGetVolumeStats implements csi.NodeServer.
func (n *NodeServer) NodeGetVolumeStats(_ context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeGetVolumeStats is not implemented")
}

// NodeExpandVolume implements csi.NodeServer.
func (n *NodeServer) NodeExpandVolume(_ context.Context, _ *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

// NodePublishVolume implements csi.NodeServer.
// Mounts the volume.
func (n *NodeServer) NodePublishVolume(_ context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	stagingTargetPath := req.GetStagingTargetPath()

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: volume capability is missing")
	}
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Volume ID missing in request")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Staging target path missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: Target path missing in request")
	}

	notMount, err := checkMount(stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if notMount {
		// Staged mount is dead by some reason. Revive it
		bucketName, prefix := volumeIDToBucketPrefix(volumeId)
		s3, err := utils.NewClientFromSecrets(req.GetSecrets())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize S3 client: %s", err)
		}
		meta := getMeta(bucketName, prefix, req.VolumeContext)
		mounter, err := mounter.NewMounter(meta, s3.Config)
		if err != nil {
			return nil, err
		}
		if err := mounter.Mount(stagingTargetPath, volumeId); err != nil {
			return nil, err
		}
	}

	// check if the volume is already being published to the target path
	notMount, err = checkMount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMount {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// TODO: Implement readOnly & mountFlags
	readOnly := req.GetReadonly()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	attrib := req.GetVolumeContext()

	klog.V(2).Infof("NodePublishVolume: volumeID %s, targetPath %s, stagingTargetPath %s, readOnly %v, mountFlags %v, attributes %v",
		volumeId, targetPath, stagingTargetPath, readOnly, mountFlags, attrib)

	cmd := exec.Command("mount", "--bind", stagingTargetPath, targetPath)
	cmd.Stderr = os.Stderr
	klog.V(4).Infof("s3: mounting volume %s to %s", volumeId, targetPath)
	out, err := cmd.Output()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "NodePublishVolume: failed to mount %s to %s: %v", stagingTargetPath, targetPath, out)
	}

	klog.V(2).Infof("NodePublishVolume: volume (%s) mounted to %s", volumeId, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume implements csi.NodeServer.
// Unmounts the volume.
func (n *NodeServer) NodeUnpublishVolume(_ context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume ID missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target path missing in request")
	}

	if err := mounter.Unmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	klog.V(4).Infof("s3: volume %s has been unmounted.", volumeID)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeStageVolume implements csi.NodeServer.
func (n *NodeServer) NodeStageVolume(_ context.Context, _ *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeStageVolume is not implemented")
}

// NodeUnstageVolume implements csi.NodeServer.
func (n *NodeServer) NodeUnstageVolume(_ context.Context, _ *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeUnstageVolume is not implemented")
}

// getMeta returns the metadata for the given bucket and prefix.
func getMeta(bucketName, prefix string, context map[string]string) *utils.Metadata {
	mountOptions := make([]string, 0)
	mountOptStr := context["options"]
	if mountOptStr != "" {
		re, _ := regexp.Compile(`([^\s"]+|"([^"\\]+|\\")*")+`)
		re2, _ := regexp.Compile(`"([^"\\]+|\\")*"`)
		re3, _ := regexp.Compile(`\\(.)`)
		for _, opt := range re.FindAll([]byte(mountOptStr), -1) {
			// Unquote options
			opt = re2.ReplaceAllFunc(opt, func(q []byte) []byte {
				return re3.ReplaceAll(q[1:len(q)-1], []byte("$1"))
			})
			mountOptions = append(mountOptions, string(opt))
		}
	}
	capacity, _ := strconv.ParseInt(context["capacity"], 10, 64)
	return &utils.Metadata{
		BucketName:    bucketName,
		Prefix:        prefix,
		Mounter:       context["mounter"],
		MountOptions:  mountOptions,
		CapacityBytes: capacity,
	}
}

// checkMount checks if the target path is mounted.
func checkMount(targetPath string) (bool, error) {
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return false, err
			}
			notMnt = true
		} else {
			return false, err
		}
	}
	return notMnt, nil
}
