package options

import "fmt"

// Validate checks Options and return a slice of found errs.
func (o *Options) Validate() []error {
	var errs []error

	errs = append(errs, o.Log.Validate()...)
	errs = append(errs, o.UnixSocket.Validate()...)
	errs = append(errs, o.ServerRunOptions.Validate()...)
	errs = append(errs, o.DriverOptions.Validate()...)
	errs = append(errs, o.Kubernetes.Validate()...)

	if o.UnixSocket.Socket == "" {
		errs = append(errs, fmt.Errorf("unix socket endpoint is empty"))
	}

	return errs
}
