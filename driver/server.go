package driver

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/keington/s3-csi-driver/driver/utils"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-11 18:15:27
 * @file: server.go
 * @description: grpc 封装
 */

type NonBlockingGRPCServer interface {
	// Start starts the gRPC server and blocks until the server exits.
	Start(endponit string, is csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, mode bool)
	// WaitForStop blocks until the server is stopped.
	Wait()
	// Stop stops the gRPC server.
	Stop()
	// ForceStop stops the gRPC server immediately.
	ForceStop()
}

// NonBlockingGRPCServerOptions contains the options for a non-blocking gRPC server.
type NonBlockingGRPCServerOptions struct {
	wg     *sync.WaitGroup
	server *grpc.Server
}

// NewNonBlockingGRPCServerOptions returns a new NonBlockingGRPCServerOptions.
func NewNonBlockingGRPCServerOptions() *NonBlockingGRPCServerOptions {
	return &NonBlockingGRPCServerOptions{}
}

// Start starts the gRPC server and blocks until the server exits.
func (s *NonBlockingGRPCServerOptions) Start(endponit string, is csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, mode bool) {
	s.wg.Add(1)

	go s.serve(endponit, is, cs, ns, mode)
}

// WaitForStop blocks until the server is stopped.
func (s *NonBlockingGRPCServerOptions) Wait() {
	s.wg.Wait()
}

// Stop stops the gRPC server.
func (s *NonBlockingGRPCServerOptions) Stop() {
	s.server.GracefulStop()
}

// ForceStop stops the gRPC server immediately.
func (s *NonBlockingGRPCServerOptions) ForceStop() {
	s.server.Stop()
}

// serve starts the gRPC server and serves the given services.
func (s *NonBlockingGRPCServerOptions) serve(endpoint string, is csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, mode bool) {
	proto, addr, err := utils.ParseEndpoint(endpoint)
	if err != nil {
		klog.Fatal(err.Error())
	}

	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			klog.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		}
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(utils.LogGRPC),
	}
	server := grpc.NewServer(opts...)
	s.server = server

	if is != nil {
		csi.RegisterIdentityServer(server, is)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}

	// Used to stop the server while running tests
	if mode {
		s.wg.Done()
		go func() {
			// make sure Serve() is called
			s.wg.Wait()
			time.Sleep(time.Millisecond * 1000)
			s.server.GracefulStop()
		}()
	}

	klog.Infof("Listening for connections on address: %#v", listener.Addr())

	err = server.Serve(listener)
	if err != nil {
		klog.Fatalf("Failed to serve grpc server: %v", err)
	}
}
