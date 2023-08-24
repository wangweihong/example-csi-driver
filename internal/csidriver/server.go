package csidriver

import (
	"github.com/wangweihong/eazycloud/pkg/grpcsvr"

	"github.com/wangweihong/example-csi-driver/internal/csidriver/config"
	"github.com/wangweihong/example-csi-driver/internal/csidriver/options"

	"github.com/wangweihong/eazycloud/pkg/log"
	"github.com/wangweihong/eazycloud/pkg/shutdown"
	"github.com/wangweihong/eazycloud/pkg/shutdown/managers/posixsignal"
)

type server struct {
	grpcServer *grpcsvr.GRPCServer
	// 控制服务关闭时处理动作, 如捕捉到信号后如何处理
	gracefulShutdown *shutdown.GracefulShutdown

	driver     *options.DriverOptions
	kubeconfig string
}

// preparedServer is a private wrapper that enforces a call of PrepareRun() before Run can be invoked.
type preparedServer struct {
	*server
}

// 创建服务器实例.
func createServer(cfg *config.Config) (*server, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	// 构建通用的grpc server服务配置
	grpcConfig, err := buildGenericGRPCServerConfig(cfg)
	if err != nil {
		return nil, err
	}

	// 补全通用服务器配置, 并生成通用服务实例
	grpcServer, err := grpcConfig.Complete().New()
	if err != nil {
		return nil, err
	}

	server := &server{
		grpcServer:       grpcServer,
		gracefulShutdown: gs,
		driver:           cfg.DriverOptions,
		kubeconfig:       cfg.Kubernetes.KubeConfig,
	}

	return server, nil
}

// 根据服务器配置应用到通用服务器配置上.
func buildGenericGRPCServerConfig(cfg *config.Config) (genericConfig *grpcsvr.GRPCConfig, lastErr error) {
	genericConfig = &grpcsvr.GRPCConfig{
		UnixSocket: cfg.UnixSocket.Socket,
		MaxMsgSize: cfg.ServerRunOptions.MaxMsgSize,
	}

	if lastErr = cfg.ServerRunOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}
	return
}

// PrepareRun prepares the server to run, by setting up the server instance.
func (s *server) PrepareRun() preparedServer {
	// 设置服务优雅退出回调处理
	s.gracefulShutdown.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		s.grpcServer.Close()
		return nil
	}))

	initRouter(s)

	return preparedServer{s}
}

func (s preparedServer) Run(stopCh <-chan struct{}) error {
	// start shutdown managers
	if err := s.gracefulShutdown.Start(); err != nil {
		log.Fatalf("start shutdown server failed: %s", err.Error())
	}

	s.grpcServer.Run()
	return nil
}
