package main

import (
	"flag"
	"os"

	"github.com/keington/s3-csi-driver/driver"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-13 21:16:01
 * @file: main.go
 * @description: s3-csi-driver main
 */

var (
	endpoint                     = flag.String("endpoint", "unix://tmp/csi.sock", "CSI endpoint")
	nodeId                       = flag.String("nodeid", "", "node id")
	mountPermissions             = flag.Uint64("mount-permissions", 0, "mounted folder permissions")
	workingMountDir              = flag.String("working-mount-dir", "/tmp", "working directory for provisioner to mount nfs shares temporarily")
	volStatsCacheExpireInMinutes = flag.Int("vol-stats-cache-expire-in-minutes", 10, "The cache expire time in minutes for volume stats cache")
)

func main() {
	flag.Parse()

	driverOptions := driver.DriverOptions{
		DriverName:                      "s3.csi.k8s.io",
		NodeID:                          *nodeId,
		EndPoint:                        *endpoint,
		MountPermissions:                *mountPermissions,
		WorkingMountDir:                 *workingMountDir,
		VolumeStatsCacheExpireInMinutes: *volStatsCacheExpireInMinutes,
	}

	// Start the driver
	driver, err := driver.NewDriver(&driverOptions)
	if err != nil {
		panic(err)
	}
	driver.Run(false)
	os.Exit(0)
}
