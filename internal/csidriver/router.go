package csidriver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned/typed/volumesnapshot/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/wangweihong/eazycloud/pkg/log"

	controllerv1 "github.com/wangweihong/example-csi-driver/internal/csidriver/controller/v1"
)

func initRouter(s *server) {
	installInterceptor(s)
	installService(s)
}

func installInterceptor(s *server) {
}

func installService(s *server) {
	var config *rest.Config
	var err error
	if s.kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", s.kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		log.Fatalf("get kubeconfig error:%v", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("generate client error:%v", err.Error())
	}

	snapshotClient, err := snapshotv1beta1.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create snapshot client: %v", err)
	}

	ids := controllerv1.NewIdentifyServer(s.driver.Name)
	ns := controllerv1.NewNodeServer(s.driver.NodeID, s.driver.MaxVolumePerNode, clientset)
	cs := controllerv1.NewControllerServer(s.driver.NodeID, clientset, snapshotClient)

	csi.RegisterIdentityServer(s.grpcServer.Server, ids)
	csi.RegisterControllerServer(s.grpcServer.Server, cs)
	csi.RegisterNodeServer(s.grpcServer.Server, ns)
}
