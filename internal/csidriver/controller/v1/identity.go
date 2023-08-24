package v1

import (
	"context"

	"github.com/prometheus/common/version"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IdentityServer struct {
	name    string
	version string
}

func NewIdentifyServer(name string) *IdentityServer {
	return &IdentityServer{
		name:    name,
		version: version.Info(),
	}
}

func (ids *IdentityServer) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest,
) (*csi.GetPluginInfoResponse, error) {
	if ids.name == "" {
		return nil, status.Error(codes.Unavailable, "driver name not configured")
	}

	if ids.version == "" {
		return nil, status.Error(codes.Unavailable, "driver version is missing")
	}

	return &csi.GetPluginInfoResponse{
		Name:          ids.name,
		VendorVersion: ids.version,
	}, nil
}

// GetPluginCapabilities return capability supported by driver
// this methods return capabilities of the plugin. Currently it reports whether the plugin has the ability of
// serving the Controller interface. The CO calls the controller interface depending on whether this method return the
// capability.
// When csi-provisioner plugin start, it will call this api.
func (ids *IdentityServer) GetPluginCapabilities(
	cts context.Context,
	req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (ids *IdentityServer) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{}, nil
}
