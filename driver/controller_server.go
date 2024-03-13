package driver

import (
	"context"

	"path"
	"strings"

	"github.com/keington/s3-csi-driver/driver/utils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	common "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 14:33:43
 * @file: controller_server.go
 * @description: 控制器服务
 */

type ControllerServer struct {
	*common.DefaultControllerServer
}

// NewControllerServiceCapability creates a new ControllerServiceCapability
func NewControllerServiceCapability(cap csi.ControllerServiceCapability_RPC_Type) *csi.ControllerServiceCapability {
	return &csi.ControllerServiceCapability{
		Type: &csi.ControllerServiceCapability_Rpc{
			Rpc: &csi.ControllerServiceCapability_RPC{
				Type: cap,
			},
		},
	}
}

// CreateVolume implements csi.ControllerServer.
func (c *ControllerServer) CreateVolume(_ context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	volumeId := req.GetName()
	params := req.GetParameters()
	bucketName := volumeId
	prefix := ""

	if err := c.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.V(2).Infof("ValidateControllerServiceRequest: invalid create volume request: %v", req)
		return nil, err
	}

	// check arguments
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume: volume ID is missing")
	}
	if params["bucket"] != "" {
		bucketName = params["bucket"]
		prefix = volumeId
		volumeId = path.Join(bucketName, prefix)
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateVolume: volume capabilities is missing")
	}

	capacityBytes := int64(req.GetCapacityRange().GetRequiredBytes())

	klog.V(4).Infof("CreateVolume: volumeId %s, capacityBytes %d", volumeId, capacityBytes)

	// create s3 client and create bucket
	client, err := utils.NewClientFromSecrets(req.GetSecrets())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateVolume: failed to initialize s3 client: %s", err.Error())
	}
	exits, err := client.IsBucketExist(bucketName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CreateVolume: failed to check if bucket %s exists: %s", bucketName, err.Error())
	}
	if !exits {
		if err = client.CreateBucket(bucketName); err != nil {
			return nil, status.Errorf(codes.Internal, "CreateVolume: failed to create bucket %s: %s", bucketName, err.Error())
		}
	}
	if prefix != "" {
		if err = client.CreateBucket(prefix); err != nil {
			return nil, status.Errorf(codes.Internal, "CreateVolume: failed to create bucket %s: %s", prefix, err.Error())
		}
	}

	klog.V(4).Infof("CreateVolume: volumeId %s, capacityBytes %d", volumeId, capacityBytes)

	context := make(map[string]string)
	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeId,
			CapacityBytes: capacityBytes,
			VolumeContext: context,
		},
	}, nil
}

// DeleteVolume implements csi.ControllerServer.
func (c *ControllerServer) DeleteVolume(_ context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeId := req.GetVolumeId()
	bucketName, prefix := volumeIDToBucketPrefix(volumeId)

	// check arguments
	if len(volumeId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "DeleteVolume: volume ID is missing")
	}
	if err := c.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		klog.V(2).Infof("ValidateControllerServiceRequest: invalid delete volume request: %v", req)
		return nil, err
	}

	klog.V(4).Infof("DeleteVolume: volumeId %s", volumeId)

	// create s3 client and delete bucket
	client, err := utils.NewClientFromSecrets(req.GetSecrets())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "DeleteVolume: failed to initialize s3 client: %s", err.Error())
	}
	var deleteErr error
	if prefix != "" {
		if err := client.DeleteBucket(bucketName); err != nil {
			deleteErr = err
		}
		klog.V(4).Infof("DeleteVolume: prefix %s is deleted", prefix)
	} else {
		if err := client.DeletePrefix(bucketName, prefix); err != nil {
			deleteErr = status.Errorf(codes.Internal, "DeleteVolume: failed to delete bucket %s: %s", prefix, err.Error())
		}
		klog.V(4).Infof("DeleteVolume: prefix %s is deleted", prefix)
	}

	if deleteErr != nil {
		return nil, deleteErr
	}
	return &csi.DeleteVolumeResponse{}, nil
}

// ValidateVolumeCapabilities implements csi.ControllerServer.
func (c *ControllerServer) ValidateVolumeCapabilities(_ context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	// check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ValidateVolumeCapabilities: volume ID is missing")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "ValidateVolumeCapabilities: volume capabilities is missing")
	}

	bucketName, _ := volumeIDToBucketPrefix(req.GetVolumeId())

	// create s3 client and check if bucket exists
	client, err := utils.NewClientFromSecrets(req.GetSecrets())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ValidateVolumeCapabilities: failed to initialize s3 client: %s", err.Error())
	}
	exists, err := client.IsBucketExist(bucketName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ValidateVolumeCapabilities: failed to check if bucket %s exists: %s", bucketName, err.Error())
	}
	if !exists {
		return nil, status.Errorf(codes.NotFound, "ValidateVolumeCapabilities: bucket %s does not exist", bucketName)
	}

	supportedAccessModes := &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	}
	for _, capability := range req.GetVolumeCapabilities() {
		if capability.GetAccessMode().Mode != supportedAccessModes.GetMode() {
			return &csi.ValidateVolumeCapabilitiesResponse{
				Message: "Only single node writer is supported",
			}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: supportedAccessModes,
				},
			},
		},
	}, nil
}

// GetCapacity implements csi.ControllerServer.
func (c *ControllerServer) GetCapacity(context.Context, *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetCapacity: not implemented")
}

// ListSnapshots implements csi.ControllerServer.
func (c *ControllerServer) ListSnapshots(context.Context, *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListSnapshots: not implemented")
}

// ListVolumes implements csi.ControllerServer.
func (c *ControllerServer) ListVolumes(context.Context, *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListVolumes: not implemented")
}

// CreateSnapshot implements csi.ControllerServer.
func (c *ControllerServer) CreateSnapshot(context.Context, *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CreateSnapshot: not implemented")
}

// DeleteSnapshot implements csi.ControllerServer.
func (c *ControllerServer) DeleteSnapshot(context.Context, *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteSnapshot: not implemented")
}

// ControllerExpandVolume implements csi.ControllerServer.
func (c *ControllerServer) ControllerExpandVolume(context.Context, *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerExpandVolume: not implemented")
}

// ControllerGetCapabilities implements csi.ControllerServer.
func (c *ControllerServer) ControllerGetCapabilities(context.Context, *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetCapabilities: not implemented")
}

// ControllerGetVolume implements csi.ControllerServer.
func (c *ControllerServer) ControllerGetVolume(context.Context, *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerGetVolume: not implemented")
}

// ControllerModifyVolume implements csi.ControllerServer.
func (c *ControllerServer) ControllerModifyVolume(context.Context, *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerModifyVolume: not implemented")
}

// ControllerPublishVolume implements csi.ControllerServer.
func (c *ControllerServer) ControllerPublishVolume(context.Context, *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerPublishVolume: not implemented")
}

// ControllerUnpublishVolume implements csi.ControllerServer.
func (c *ControllerServer) ControllerUnpublishVolume(context.Context, *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerUnpublishVolume: not implemented")
}

// volumeIDToBucketPrefix returns the bucket name and prefix based on the volumeID.
// Prefix is empty if volumeID does not have a slash in the name.
func volumeIDToBucketPrefix(volumeID string) (string, string) {
	// if the volumeID has a slash in it, this volume is
	// stored under a certain prefix within the bucket.
	splitVolumeID := strings.SplitN(volumeID, "/", 2)
	if len(splitVolumeID) > 1 {
		return splitVolumeID[0], splitVolumeID[1]
	}

	return volumeID, ""
}
