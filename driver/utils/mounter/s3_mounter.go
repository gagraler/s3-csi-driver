package mounter

import (
	"fmt"
	"os"

	"github.com/keington/s3-csi-driver/driver/utils"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-13 21:14:08
 * @file: s3_mounter.go
 * @description: s3挂载器
 */

type S3Mounter struct {
	meta          *utils.Metadata
	url           string
	region        string
	pwFileContent string
}

// NewS3Mounter creates a new S3 fs mounter
func NewS3Mounter(meta *utils.Metadata, cfg *utils.Config) (Mounter, error) {
	return &S3Mounter{
		meta:          meta,
		url:           cfg.Endpoint,
		region:        cfg.Region,
		pwFileContent: cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

// Mount mounts the s3fs
func (s *S3Mounter) Mount(target, volumeID string) error {

	if err := writeS3Pass(s.pwFileContent); err != nil {
		return err
	}
	args := []string{
		fmt.Sprintf("%s:%s", s.meta.BucketName, s.meta.Prefix),
		target,
		"-o", "use_path_request_style",
		"-o", fmt.Sprintf("url=%s", s.url),
		"-o", "allow_other",
		"-o", "mp_umask=000",
	}
	if s.region != "" {
		args = append(args, "-o", fmt.Sprintf("endpoint=%s", s.region))
	}
	return FuseMount(target, "s3fs", args, nil)
}

// writeS3Pass writes the s3 password to a file
func writeS3Pass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-s3", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
