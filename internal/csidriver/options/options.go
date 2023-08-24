package options

import (
	"github.com/wangweihong/eazycloud/pkg/app"
	cliflag "github.com/wangweihong/eazycloud/pkg/cli/flag"
	"github.com/wangweihong/eazycloud/pkg/grpcsvr/grpcoptions"
	"github.com/wangweihong/eazycloud/pkg/json"
	"github.com/wangweihong/eazycloud/pkg/log"
)

var (
	_ app.PrintableOptions    = &Options{}
	_ app.CompleteableOptions = &Options{}
)

// Options runs a http server.
type Options struct {
	Log              *log.Options                   `json:"log"        mapstructure:"log"`
	UnixSocket       *grpcoptions.UnixSocketOptions `json:"unix"       mapstructure:"unix"`
	ServerRunOptions *grpcoptions.ServerRunOptions  `json:"server"     mapstructure:"server"`
	DriverOptions    *DriverOptions                 `json:"driver"     mapstructure:"driver"`
	Kubernetes       *KubernetesOptions             `json:"kubernetes" mapstructure:"kubernetes"`
}

// NewOptions creates a new Options object with default parameters.
func NewOptions() *Options {
	s := Options{
		Log:              log.NewOptions(),
		UnixSocket:       grpcoptions.NewUnixSocketOptions(),
		ServerRunOptions: grpcoptions.NewServerRunOptions(),
		DriverOptions:    NewDriverOptions(),
		Kubernetes:       NewKubernetesOptions(),
	}

	return &s
}

// Flags returns flags for a specific server by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.Log.AddFlags(fss.FlagSet("logs"))
	o.UnixSocket.AddFlags(fss.FlagSet("unix"))
	o.ServerRunOptions.AddFlags(fss.FlagSet("server"))
	o.DriverOptions.AddFlags(fss.FlagSet("driver"))
	o.Kubernetes.AddFlags(fss.FlagSet("kubernetes"))

	return fss
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}

// Complete fills in any fields not set that are required to have valid data.
// 补全指定的选项.
func (o *Options) Complete() error {
	return nil
}
