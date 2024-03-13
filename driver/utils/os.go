package utils

import (
	"os"
	"k8s.io/klog/v2"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-12 17:02:10
 * @file: os.go
 * @description: os工具
 */

// ChmodIfPermissionMismatch checks the permission of the targetPath and chmod it if it's different from the mode.
func ChmodIfPermissionMismatch(targetPath string, mode os.FileMode) error {
	info, err := os.Lstat(targetPath)
	if err != nil {
		return err
	} 
	
	perm := info.Mode() & os.ModePerm
	if perm != mode {
		klog.V(2).Infof("chmod targetPath(%s, mode:0%o) with permissions(0%o)", targetPath, info.Mode(), mode)
		if err := os.Chmod(targetPath, mode); err != nil {
			return err
		}
	} else {
		klog.V(2).Infof("skip chmod on targetPath(%s) since mode is already 0%o)", targetPath, info.Mode())
	}
	return nil
}