package pkg

import (
	"fmt"
	"runtime"
	"strings"

	"sigs.k8s.io/yaml"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 17:34:48
 * @file: version.go
 * @description: version info
 */

var (
	DriverVersion = "null"
	gitCommit     = "null"
	buildDate     = "null"
)

type VersionInfo struct {
	DriverName    string `json:"Driver Name"`
	DriverVersion string `json:"Driver Version"`
	GitCommit     string `json:"Git Commit"`
	BuildDate     string `json:"Build Date"`
	GoVersion     string `json:"Go Version"`
	Compiler      string `json:"Compiler"`
	Platform      string `json:"Platform"`
}

// GetVersionInfo returns the version information of the driver.
func GetVersionInfo(driverName string) *VersionInfo {
	return &VersionInfo{
		DriverName:    driverName,
		DriverVersion: DriverVersion,
		GitCommit:     gitCommit,
		BuildDate:     buildDate,
		GoVersion:     runtime.Version(),
		Compiler:      runtime.Compiler,
		Platform:      fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// GetVersionYaml returns the version information of the driver in YAML format.
func GetVersionYaml(driverName string) (string, error) {
	info := GetVersionInfo(driverName)
	marshal, err := yaml.Marshal(&info)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(marshal)), nil
}
