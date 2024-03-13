package utils

import (
	"context"
	"fmt"
	"strings"
	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 18:24:36
 * @file: grpc.go
 * @description: grpc utils
 */

// ParseEndpoint parses the given endpoint and returns the protocol and address.
// It supports both "unix://" and "tcp://" protocols.
// If the endpoint is invalid, it returns an error.
func ParseEndpoint(endpoint string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(endpoint), "unix://") || strings.HasPrefix(strings.ToLower(endpoint), "tcp://") {
		s := strings.SplitN(endpoint, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
	}
	return "", "", fmt.Errorf("invalid endpoint: %v", endpoint)
}

// LogGRPC is a middleware function that logs the details of a gRPC call.
// It logs the method, request, and response using the klog package.
// The log level is determined based on the method.
// It returns the response and any error from the handler.
func LogGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	level := klog.Level(getLogLevel(info.FullMethod))
	klog.V(level).Infof("GRPC call: %s", info.FullMethod)
	klog.V(level).Infof("GRPC request: %s", protosanitizer.StripSecrets(req))

	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.V(level).Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}

// getLogLevel returns the log level based on the given method.
// Certain methods have a higher log level for more detailed logging.
func getLogLevel(method string) int32 {
	if method == "/csi.v1.Identity/Probe" ||
		method == "/csi.v1.Node/NodeGetCapabilities" ||
		method == "/csi.v1.Node/NodeGetVolumeStats" {
		return 8
	}
	return 2
}