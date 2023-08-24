package options

import "github.com/spf13/pflag"

// Options describe how to connect to kubernetes.
type KubernetesOptions struct {
	KubeConfig string `json:"kubeconfig" mapstructure:"kubeconfig"`
}

// NewOptions creates a new Options object with default parameters.
func NewKubernetesOptions() *KubernetesOptions {
	s := KubernetesOptions{}

	return &s
}

func (o *KubernetesOptions) Validate() []error {
	var errs []error

	return errs
}

// AddFlags adds flags related to features for kubernetes to the
// specified FlagSet.
func (o *KubernetesOptions) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		return
	}

	fs.StringVar(&o.KubeConfig, "kubernetes.kubeconfig", o.KubeConfig, ""+
		"path of kubernetes kubeconfig file. Required when run outside of kubernetes cluster")
}
