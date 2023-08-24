package storage

import (
	"context"
	"os"
)

type VolumeAPI interface {
	Create(
		ctx context.Context,
		name string,
		sizeBytes int64,
		snapshotID string,
		parameters map[string]string,
	) (string, error)
	Delete(ctx context.Context, parameters map[string]string) error
	Expand(
		ctx context.Context,
		name string,
		parameters map[string]string,
		size int64,
	) (nodeExpansionRequired bool, err error)
	Attach(ctx context.Context, name string, parameters map[string]interface{}) error
	Detach(ctx context.Context, name string, parameters map[string]interface{}) error
	Stage(ctx context.Context, name string, parameters map[string]string, mountPerm os.FileMode) error
	Unstage(ctx context.Context, name string, parameters map[string]string) error
	NodeExpand(ctx context.Context, name string, parameters map[string]string) error
}
