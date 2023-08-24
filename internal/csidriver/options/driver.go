package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// Options runs a csi driver.
type DriverOptions struct {
	// Endpoint string `json:"endpoint"   mapstructure:"endpoint"`
	Name             string `json:"name"       mapstructure:"name"`
	NodeID           string `json:"node-id"    mapstructure:"node-id"`
	MaxVolumePerNode int    `json:"max-volume" mapstructure:"max-volume"`
	RunAsController  bool   `json:"controller" mapstructure:"controller"`
}

// NewOptions creates a new Options object with default parameters.
func NewDriverOptions() *DriverOptions {
	s := DriverOptions{
		Name:            "example-csi-driver",
		RunAsController: false,
	}

	return &s
}

func (o *DriverOptions) Validate() []error {
	var errs []error

	if o.NodeID == "" {
		errs = append(errs, fmt.Errorf("driver node id is empty"))
	}
	return errs
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (o *DriverOptions) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		return
	}

	fs.StringVar(&o.Name, "driver.name", o.Name, ""+
		"Name of csi driver.")
	fs.StringVar(&o.NodeID, "driver.node-id", o.NodeID, ""+
		"Node id of csi driver.")
	fs.IntVar(&o.MaxVolumePerNode, "driver.max-volume", o.MaxVolumePerNode, ""+
		"Max volume per node")
	fs.BoolVar(&o.RunAsController, "driver.controller", o.RunAsController, ""+
		"This driver run as a controller service. if false, driver will run as agent service.")
}
