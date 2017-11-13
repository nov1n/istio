package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/grpc"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"istio.io/istio/pilot/cmd"
	"istio.io/istio/pilot/proxy/envoy/v2"
)

func main() {
	flag.Parse()
	stop := make(chan struct{})

	if namespace == "" {
		namespace = os.Getenv("POD_NAMESPACE")
	}

	config := cache.NewSimpleCache(v2.Hasher{}, nil /* TODO */)
	server := xds.NewServer(config)
	grpcServer := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	server.Register(grpcServer)

	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			glog.Error(err)
		}
	}()

	generator := &v2.Generator{}
	generator.Cache = config
	generator.Generate()

	cmd.WaitSignal(stop)
}

var (
	kubeconfig string
	namespace  string
	port       int

	validate bool
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Use a Kubernetes configuration file instead of in-cluster configuration")
	flag.StringVar(&namespace, "namespace", "",
		"Select a namespace where the controller resides. If not set, uses ${POD_NAMESPACE} environment variable")
	flag.IntVar(&port, "port", 15003,
		"ADS port")
}