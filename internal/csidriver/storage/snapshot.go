package storage

import (
	"context"
	"fmt"
	"strings"
)

type SnapshotAPI interface {
	Create(
		ctx context.Context,
		name string,
		snapshotID string,
		parameters map[string]string,
	) (map[string]interface{}, error)
	Delete(ctx context.Context, volumeSnapshotID string, parameters map[string]string) error
}

type VolumeSnapshotInfo struct {
	Name     string
	ID       string
	VolumeID string
}

func (vsi VolumeSnapshotInfo) String() string {
	return strings.Join([]string{vsi.Name, vsi.ID, vsi.VolumeID}, ".")
}

func ConvertVolumeSnapshotInfoFromSnapshotHandle(csiSnapshotID string) (*VolumeSnapshotInfo, error) {
	volumeSlice := strings.Split(csiSnapshotID, ".")
	if len(volumeSlice) < 3 {
		return nil, fmt.Errorf("invalid  csi snapshot id: " + csiSnapshotID)
	}
	vi := &VolumeSnapshotInfo{
		Name:     volumeSlice[0],
		ID:       volumeSlice[1],
		VolumeID: volumeSlice[2],
	}
	return vi, nil
}
