package driver

import (
	"strings"

	"github.com/keington/s3-csi-driver/driver/pkg"
	"github.com/keington/s3-csi-driver/driver/utils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	common "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 10:27:53
 * @file: driver.go
 * @description: driver info
 */

const (
	// DefaultDriverName is the default name of the driver.
	DefaultDriverName = "minio.csi.k8s.io"
	// ParamServer is the address of the minio server.
	ParamServer = "server"
	// ParamShare is the base directory of the NFS server to create volumes under.
	ParamShare = "share"
	// ParamSubDir is the subdirectory under the base directory to create volumes under.
	ParamSubDir = "subdir"
	// ParamOnDelete is the policy for handling volumes on delete.
	ParamOnDelete = "ondelete"
	// MountOptionsField is the field name for mount options.
	MountOptionsField = "mountoptions"
	// MountPermissionsField is the field name for mount permissions.
	MountPermissionsField = "mountpermissions"
	// PvcNameKey is the key for PVC name.
	PvcNameKey = "csi.storage.k8s.io/pvc/name"
	// PvcNamespaceKey is the key for PVC namespace.
	PvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	// PvNameKey is the key for PV name.
	PvNameKey = "csi.storage.k8s.io/pv/name"
	// PvcNameMetadata is the metadata for PVC name.
	PvcNameMetadata = "${pvc.metadata.name}"
	// PvcNamespaceMetadata is the metadata for PVC namespace.
	PvcNamespaceMetadata = "${pvc.metadata.namespace}"
	// PvNameMetadata is the metadata for PV name.
	PvNameMetadata = "${pv.metadata.name}"
)

// Driver represents the CSI driver.
type Driver struct {
	Name                            string                             // Name is the name of the driver.
	NodeID                          string                             // NodeID is the ID of the node where the driver is running.
	Version                         string                             // Version is the version of the driver.
	EndPoint                        string                             // EndPoint is the endpoint of the driver.
	MountPermissions                uint64                             // MountPermissions is the mount permissions for the driver.
	WorkingMountDir                 string                             // WorkingMountDir is the working directory for mount operations.
	DefaultOnDeletePolicy           string                             // DefaultOnDeletePolicy is the default policy for handling volumes on delete.
	NodeServer                      *NodeServer                        // NodeServer is the server for handling node service requests.
	ControllerServer                *ControllerServer                  // ControllerServer is the server for handling controller service requests.
	IdentityServer                  *IdentityServer                    // IdentityServer is the server for handling identity service requests.
	Driver                          *common.CSIDriver                  // Driver is the CSI driver.
	ControllerServiceCapability     []*csi.ControllerServiceCapability // ControllerServiceCapability represents the capabilities of the controller service.
	NodeServiceCapability           []*csi.NodeServiceCapability       // NodeServiceCapability represents the capabilities of the node service.
	VolumeLocks                     *utils.VolumeLocks                 // VolumeLocks is used for locking volumes during operations.
	VolumeStatsCache                cache.Resource                     // VolumeStatsCache is the cache for volume statistics.
	VolumeStatsCacheExpireInMinutes int                                // VolumeStatsCacheExpireInMinutes is the expiration time for volume statistics cache.
}

// DriverOptions represents the options for creating a new driver.
type DriverOptions struct {
	DriverName                      string // DriverName is the name of the CSI driver.
	NodeID                          string // NodeID is the unique identifier of the node where the driver is running.
	EndPoint                        string // EndPoint is the CSI endpoint address.
	MountPermissions                uint64 // MountPermissions is the permission mode for mounting volumes.
	WorkingMountDir                 string // WorkingMountDir is the directory where volumes are mounted.
	VolumeStatsCacheExpireInMinutes int    // VolumeStatsCacheExpireInMinutes is the expiration time for volume statistics cache in minutes.
}

// NewDriver creates a new driver object.
func NewDriver(options *DriverOptions) (*Driver, error) {
	klog.Infof("driver: %v version: %v", options.DriverName, pkg.DriverVersion)

	driver := &Driver{
		Name:                            options.DriverName,
		Version:                         pkg.DriverVersion,
		NodeID:                          options.NodeID,
		MountPermissions:                options.MountPermissions,
		WorkingMountDir:                 options.WorkingMountDir,
		VolumeStatsCacheExpireInMinutes: options.VolumeStatsCacheExpireInMinutes,
	}

	return driver, nil
}

// NewControllerServer creates a new controller server.
func NewControllerServer(d *common.CSIDriver) *ControllerServer {
	return &ControllerServer{
		DefaultControllerServer: common.NewDefaultControllerServer(d),
	}
}

// NewNodeServer creates a new node server.
func NewNodeServer(d *common.CSIDriver) *NodeServer {
	return &NodeServer{
		DefaultNodeServer: common.NewDefaultNodeServer(d),
	}
}

// NewIdentityServer creates a new identity server.
func NewIdentityServer(d *common.CSIDriver) *IdentityServer {
	return &IdentityServer{
		DefaultIdentityServer: common.NewDefaultIdentityServer(d),
	}
}

// Run starts the driver in the specified mode.
func (d *Driver) Run(mode bool) {
	versionMeta, err := pkg.GetVersionYaml(d.Name)
	if err != nil {
		klog.Fatalf("Failed to get version info: %v", err)
	}
	klog.Infof("Driver infomation meta: %s", versionMeta)

	// Create a new CSI driver.
	d.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
	})
	d.Driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
	})

	// Create gRPC servers.
	d.ControllerServer = NewControllerServer(d.Driver)
	d.NodeServer = NewNodeServer(d.Driver)
	d.IdentityServer = NewIdentityServer(d.Driver)

	// Start the gRPC servers.
	s := NewNonBlockingGRPCServerOptions()
	s.Start(d.EndPoint, d.IdentityServer, d.ControllerServer, d.NodeServer, mode)
	s.Wait()
}

// AddControllerServiceCapabilities adds the given controller service capabilities to the driver.
func (d *Driver) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) {
	var csc []*csi.ControllerServiceCapability
	for _, c := range cl {
		csc = append(csc, NewControllerServiceCapability(c))
	}
	d.ControllerServiceCapability = csc
}

// IsCorruptDir checks if the directory is corrupt.
// 检查目录是否损坏
func IsCorruptDir(dir string) bool {
	_, err := mount.PathExists(dir)
	return err != nil && mount.IsCorruptedMnt(err)
}

// replaceWithMap replaces the keys in the string with the values from the map.
func replaceWithMap(s string, m map[string]string) string {
	for k, v := range m {
		if k != "" {
			s = strings.ReplaceAll(s, k, v)
		}
	}

	return s
}
