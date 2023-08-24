package csidriver

import (
	"github.com/wangweihong/eazycloud/pkg/app"
	"github.com/wangweihong/eazycloud/pkg/log"

	"github.com/wangweihong/example-csi-driver/internal/csidriver/config"
	"github.com/wangweihong/example-csi-driver/internal/csidriver/options"
)

const commandDesc = `Example of csi driver`

// NewApp creates an App object with default parameters.
func NewApp(basename string) *app.App {
	// 设置应用默认参数, 并绑定对应的标志
	opts := options.NewOptions()

	// 初始化应用实例, 解析参数、绑定标志等
	application := app.NewApp("example-csi-driver",
		basename,                         // 应用名, 该名字将在未指定配置文件名时,作为默认配置文件名
		app.WithOptions(opts),            // 设置应用参数
		app.WithDescription(commandDesc), // 设置应用描述
		app.WithDefaultValidArgs(),       // 设置应用命令检测参数. 默认是应用不能带有命令
		app.WithRunFunc(run(opts)),       // 设置应用运行方法
		// app.WithNoConfig(),               // 指明应用不需要配置文件
	)

	return application
}

// 应用运行逻辑.
func run(opts *options.Options) app.RunFunc {
	return func(basename string) error {
		log.Init(opts.Log)
		defer log.Flush()

		cfg, err := config.CreateConfigFromOptions(opts)
		if err != nil {
			return err
		}

		stopCh := make(chan struct{})

		return Run(cfg, stopCh)
	}
}
