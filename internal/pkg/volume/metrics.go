package volume

import (
	"fmt"

	"golang.org/x/sys/unix"

	"k8s.io/apimachinery/pkg/api/resource"
)

type Metrics struct {
	Available  *resource.Quantity
	Capacity   *resource.Quantity
	InodesUsed *resource.Quantity
	Inodes     *resource.Quantity
	InodesFree *resource.Quantity
	Used       *resource.Quantity
}

func GetPathMetrics(path string) (*Metrics, error) {
	volumeMetrics := &Metrics{}

	inodes, inodesFree, inodesUsed, available, capacity, usage, err := FsInfo(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get FsInfo, error %w", err)
	}
	volumeMetrics.Inodes = resource.NewQuantity(inodes, resource.BinarySI)
	volumeMetrics.InodesFree = resource.NewQuantity(inodesFree, resource.BinarySI)
	volumeMetrics.InodesUsed = resource.NewQuantity(inodesUsed, resource.BinarySI)
	volumeMetrics.Available = resource.NewQuantity(available, resource.BinarySI)
	volumeMetrics.Capacity = resource.NewQuantity(capacity, resource.BinarySI)
	volumeMetrics.Used = resource.NewQuantity(usage, resource.BinarySI)

	return volumeMetrics, nil
}

func FsInfo(path string) (int64, int64, int64, int64, int64, int64, error) {
	statfs := &unix.Statfs_t{}
	err := unix.Statfs(path, statfs)
	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	capacity := int64(statfs.Blocks) * statfs.Bsize
	available := int64(statfs.Bavail) * statfs.Bsize
	used := (int64(statfs.Blocks) - int64(statfs.Bfree)) * statfs.Bsize

	inodes := int64(statfs.Files)
	inodesFree := int64(statfs.Ffree)
	inodesUsed := inodes - inodesFree
	return inodes, inodesFree, inodesUsed, available, capacity, used, nil
}
