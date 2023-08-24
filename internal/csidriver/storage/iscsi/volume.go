package iscsi

import (
	"context"
	"os"
)

type volumer struct {
	c *client
}

func newVolumer(c *client) *volumer {
	return &volumer{c: c}
}

func (c *volumer) Create(
	ctx context.Context,
	name string,
	sizeBytes int64,
	snapshotID string,
	parameters map[string]string,
) (string, error) {
	return "", nil
}
func (c *volumer) Delete(ctx context.Context, parameters map[string]string) error { return nil }
func (c *volumer) Expand(
	ctx context.Context,
	name string,
	parameters map[string]string,
	size int64,
) (nodeExpansionRequired bool, err error) {
	return false, nil
}

func (c *volumer) Attach(ctx context.Context, name string, parameters map[string]interface{}) error {
	return nil
}

func (c *volumer) Detach(ctx context.Context, name string, parameters map[string]interface{}) error {
	return nil
}

func (c *volumer) Stage(ctx context.Context, name string, parameters map[string]string, mountPerm os.FileMode) error {
	return nil
}

func (c *volumer) Unstage(ctx context.Context, name string, parameters map[string]string) error {
	return nil
}

func (c *volumer) NodeExpand(ctx context.Context, name string, parameters map[string]string) error {
	return nil
}
