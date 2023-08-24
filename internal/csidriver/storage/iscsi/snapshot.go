package iscsi

import "context"

type snapshotter struct {
	c *client
}

func newSnapshotter(c *client) *snapshotter {
	return &snapshotter{c: c}
}

func (c *snapshotter) Create(
	ctx context.Context,
	name string,
	snapshotID string,
	parameters map[string]string,
) (map[string]interface{}, error) {
	return nil, nil
}

func (c *snapshotter) Delete(ctx context.Context, volumeSnapshotID string, parameters map[string]string) error {
	return nil
}
