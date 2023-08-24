package iscsi

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/wangweihong/eazycloud/pkg/httpcli"

	"github.com/wangweihong/example-csi-driver/internal/csidriver/storage"
)

type client struct {
	*httpcli.Client

	timeout time.Duration
}

func (c *client) Volumes() storage.VolumeAPI {
	return newVolumer(c)
}

func (c *client) Snapshots() storage.SnapshotAPI {
	return newSnapshotter(c)
}

var (
	iscsiFactory storage.Factory
	once         sync.Once
)

// GetFactoryOr create iscsi factory with the given config.
func GetFactoryOr(opts *storage.StorageOption) (storage.Factory, error) {
	if opts == nil || opts.Secret == nil || opts.Parameters == nil {
		return nil, fmt.Errorf("invalid storage options")
	}

	var c *httpcli.Client
	var err error
	once.Do(func() {
		// 建立长连接?
		HTTPTransport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // 连接超时时间
				KeepAlive: 60 * time.Second, // 保持长连接的时间
			}).DialContext, // 设置连接的参数
			MaxIdleConns:          500,              // 最大空闲连接
			IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
			ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
			MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
		}
		hc := &client{
			timeout: 30 * time.Second,
		}

		c, err = httpcli.NewClient(
			opts.Address,
			httpcli.WithTransport(HTTPTransport),
			httpcli.WithTimeout(30*time.Second),
			httpcli.WithIntercepts(
			//// 注意顺序, 队列也靠后的越早执行调用后
			//interceptorcli.TokenInterceptor("TokenInterceptor", hc, skipper.AllowPathPrefixSkipper("/gettoken")),
			//ErrorCodeInterceptor("ErrorCodeInterceptor"),
			//NoSuccessStatusCodeInterceptor("NoSuccessStatusCodeInterceptor"),
			//LoggingInterceptor("Logging"),
			),
		)
		hc.Client = c

		iscsiFactory = hc
	})

	if iscsiFactory == nil || err != nil {
		return nil, fmt.Errorf("failed to get  factory, iscsiFactory: %+v, error: %w", iscsiFactory, err)
	}

	return iscsiFactory, nil
}
