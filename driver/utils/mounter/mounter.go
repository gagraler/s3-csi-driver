package mounter

import (
	"os"
	"os/exec"
	"time"

	"github.com/keington/s3-csi-driver/driver/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
	mount "k8s.io/mount-utils"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-13 21:14:00
 * @file: mounter.go
 * @description: 挂载器
 */

// Mounter interface which can be implemented
// by the different mounter types
type Mounter interface {
	Mount(target, volumeID string) error
}

// NewMounter creates a new mounter
func NewMounter(meta *utils.Metadata, cfg *utils.Config) (Mounter, error) {
	mounterType := meta.Mounter
	if len(mounterType) == 0 {
		mounterType = cfg.Mounter
	}
	switch mounterType {
	case "s3fs":
		return NewS3Mounter(meta, cfg)
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Mounter %s not supported", mounterType)
	}
}

// Unmount unmounts the volume
func Unmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	return nil
}

// FuseMount mounts the fuse
func FuseMount(path string, command string, args []string, envs []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	// cmd.Environ() returns envs inherited from the current process
	cmd.Env = append(cmd.Environ(), envs...)
	klog.V(3).Infof("Mounting fuse with command: %s and args: %s", command, args)

	out, err := cmd.Output()
	if err != nil {
		return status.Errorf(codes.Internal, "Mount: failed to mount %s: %v", path, out)
	}

	return waitForMount(path, 10*time.Second)
}

// waitForMount waits for the mount to be ready
// before returning
func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.New("").IsLikelyNotMountPoint(path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return status.Errorf(codes.Internal, "Mount: failed to mount %s: timeout", path)
		}
	}
}
