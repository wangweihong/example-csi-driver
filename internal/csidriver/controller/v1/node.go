package v1

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/wangweihong/example-csi-driver/internal/pkg/volume"
	"github.com/wangweihong/example-csi-driver/third_party/k8s.io/mount"

	"github.com/wangweihong/example-csi-driver/internal/csidriver/paramkey"
	"github.com/wangweihong/example-csi-driver/internal/csidriver/storage"

	"github.com/wangweihong/eazycloud/pkg/util/maputil"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/wangweihong/eazycloud/pkg/log"
	//	"github.com/wangweihong/eazycloud/pkg/util/diskutil"
)

const (
	TopologyKeyNode = "topology.example.csi/node"

	DefaultFileSystemType = "ext4"
	DefaultFileMode       = 0o744
)

var _ csi.NodeServer = (*nodeServer)(nil)

type nodeServer struct {
	nodeID            string
	maxVolumesPerNode int
	client            *kubernetes.Clientset
}

func NewNodeServer(nodeID string, maxVolumesPerNode int, clientSet *kubernetes.Clientset) csi.NodeServer {
	return &nodeServer{
		nodeID:            nodeID,
		maxVolumesPerNode: maxVolumesPerNode,
		client:            clientSet,
	}
}

// NodeStageVolume: mount volume to host staging path
// staging path: 通常指的是节点（Node）驱动程序（Driver）中用于临时存储挂载卷（Volume）数据的路径。该路径位于节点的本地文件系统上，用于暂存卷数据的中间过程。
func (n nodeServer) NodeStageVolume(
	ctx context.Context,
	request *csi.NodeStageVolumeRequest,
) (*csi.NodeStageVolumeResponse, error) {
	if request.GetVolumeId() == "" {
		log.F(ctx).Error("volumeID is missing")
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if request.GetStagingTargetPath() == "" {
		log.F(ctx).Error("stagingTargetPath is missing")
		return nil, status.Error(codes.InvalidArgument, "Staging Target path missing in request")
	}

	if request.GetVolumeCapability() == nil {
		log.F(ctx).Error("volume capability is missing")
		return nil, status.Error(codes.InvalidArgument, "Volume Capability missing in request")
	}

	if request.GetVolumeContext() == nil {
		log.F(ctx).Error("volume context is missing")
		return nil, status.Error(codes.InvalidArgument, "Volume Context missing in request")
	}

	// pv id
	volumeId := request.GetVolumeId()
	// host path volume will mount
	// for example: /var/lib/kubelet/plugins/kubernetes.io/csi/pv/pvc-c271aab5-3d42-4b88-adfb-e19734e8153e/globalmount
	stagingTargetPath := request.GetStagingTargetPath()
	// parameters come from storgeclass.parameters( it will remove some preset key like ‘csi.storage.k8s.io/fstype’）
	parameters := request.GetVolumeContext()
	parameters[paramkey.BackendStorageVolumeStagePath] = stagingTargetPath
	// accessMode come from pvc'accessModes
	accessMode := volume.ConvertCSIAccessMode(request.GetVolumeCapability().GetAccessMode().Mode)
	mnt := request.GetVolumeCapability().GetMount()

	// this parameter comm fro
	opts := mnt.GetMountFlags()
	if accessMode == volume.AccessModeReadOnly {
		opts = append(opts, "ro")
	}

	parameters[paramkey.BackendStorageVolumeMountFlag] = strings.Join(opts, ",")

	// set default fs type is ext4
	fsType := DefaultFileSystemType
	// this parameter come from storageclass.parameters['csi.storage.k8s.io/fstype']
	if mnt.GetFsType() != "" {
		fsType = mnt.FsType
	}
	mountPerm := os.FileMode(DefaultFileMode)
	parameters[paramkey.BackendStorageVolumeFilesystemType] = fsType
	if mountPermString, ok := parameters[paramkey.BackendStorageVolumeMountPermission]; ok {
		var err error
		uint64data, err := strconv.ParseUint(mountPermString, 0, 32)
		if err != nil {
			log.F(ctx).Errorf("incorrect mount perm param:%v", err.Error())
			return nil, status.Error(codes.InvalidArgument, "mount_perm is not correct file mode:"+mountPermString)
		}
		mountPerm = os.FileMode(uint64data)
	}

	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Errorf("select storage service from parameter error:%v", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := storagePlugin.Volumes().Stage(ctx, volumeId, parameters, mountPerm); err != nil {
		log.F(ctx).Errorf("stage volume err:%v", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeStageVolumeResponse{}, nil
}

// NodeUnstageVolume umount volume from host's staging path.
func (n nodeServer) NodeUnstageVolume(
	ctx context.Context,
	request *csi.NodeUnstageVolumeRequest,
) (*csi.NodeUnstageVolumeResponse, error) {
	volumeId := request.GetVolumeId()
	// host path volume will mount
	targetPath := request.GetStagingTargetPath()

	if strings.TrimSpace(targetPath) == "" {
		log.F(ctx).Error("missing target path")
		return nil, status.Error(codes.InvalidArgument, "staging Target path missing in request")
	}

	pvInfo, err := n.client.CoreV1().PersistentVolumes().Get(ctx, volumeId, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).Error("get pv info error", log.Err(err))
		return nil, status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("NodeUnstageVolume: get pv %v err: %s", volumeId, err.Error()),
		)
	}
	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not CSI", log.Any("pv", pvInfo))
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("NodeUnstageVolume: volume %v is no csi", volumeId))
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	parameters[paramkey.BackendStorageVolumeStagePath] = targetPath

	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Errorf("select storage service from parameter error:%v", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := storagePlugin.Volumes().Unstage(ctx, volumeId, parameters); err != nil {
		log.F(ctx).Errorf("unstage volume error:%v", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeUnpublishVolume mount volume to pod path.
func (n nodeServer) NodePublishVolume(
	ctx context.Context,
	request *csi.NodePublishVolumeRequest,
) (*csi.NodePublishVolumeResponse, error) {
	// host path volume mounted.
	stagingTargetPath := request.GetStagingTargetPath()
	// target pod path
	targetPath := request.GetTargetPath()

	if strings.TrimSpace(stagingTargetPath) == "" {
		log.F(ctx).Error("stagingTargetPath is empty")
		return nil, status.Error(codes.InvalidArgument, "staging Target path missing in request")
	}

	if strings.TrimSpace(targetPath) == "" {
		log.F(ctx).Error("targetPath is empty")
		return nil, status.Error(codes.InvalidArgument, "publish Target path missing in request")
	}

	opts := []string{"bind"}
	if request.GetReadonly() {
		opts = append(opts, "ro")
	}

	mounter := mount.New("")

	if err := mounter.Mount(stagingTargetPath, targetPath, "", opts); err != nil {
		msg := "bind mount error:" + err.Error()
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume umount volume from pod path.
func (n nodeServer) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest,
) (*csi.NodeUnpublishVolumeResponse, error) {
	if req.GetTargetPath() == "" {
		log.F(ctx).Error("target Path is empty")
		return nil, status.Error(codes.InvalidArgument, "target path missing in request")
	}

	targetPath := req.GetTargetPath()

	mounter := mount.New("")
	isMounted, err := mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		log.F(ctx).Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if isMounted {
		if err := mounter.Unmount(targetPath); err != nil {
			msg := fmt.Sprintf("umount %v error:%v", targetPath, err)
			log.F(ctx).Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (n nodeServer) NodeGetVolumeStats(
	ctx context.Context,
	req *csi.NodeGetVolumeStatsRequest,
) (*csi.NodeGetVolumeStatsResponse, error) {
	if req.GetVolumePath() == "" {
		log.F(ctx).Error("request volume path missing")
		return nil, status.Error(codes.InvalidArgument, "request volume missing")
	}

	volumeMetrics, err := volume.GetPathMetrics(req.GetVolumePath())
	if err != nil {
		msg := fmt.Sprintf("get volume metrics failed :%v", err.Error())
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	volumeAvailable, ok := volumeMetrics.Available.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics available %v is invalid", volumeMetrics.Available)
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	volumeCapacity, ok := volumeMetrics.Capacity.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics capacity %v is invalid", volumeMetrics.Capacity)
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	volumeUsed, ok := volumeMetrics.Used.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics used %v is invalid", volumeMetrics.Used)
		log.F(ctx).Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	volumeInodesFree, ok := volumeMetrics.InodesFree.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics inodesFree %v is invalid", volumeMetrics.InodesFree)
		log.F(ctx).Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	volumeInodes, ok := volumeMetrics.Inodes.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics inodes %v is invalid", volumeMetrics.Inodes)
		log.F(ctx).Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	volumeInodesUsed, ok := volumeMetrics.InodesUsed.AsInt64()
	if !ok {
		msg := fmt.Sprintf("Volume metrics inodesUsed %v is invalid", volumeMetrics.InodesUsed)
		log.F(ctx).Error(msg)
		return nil, status.Errorf(codes.Internal, msg)
	}

	response := &csi.NodeGetVolumeStatsResponse{
		Usage: []*csi.VolumeUsage{
			{
				Available: volumeAvailable,
				Total:     volumeCapacity,
				Used:      volumeUsed,
				Unit:      csi.VolumeUsage_BYTES,
			},
			{
				Available: volumeInodesFree,
				Total:     volumeInodes,
				Used:      volumeInodesUsed,
				Unit:      csi.VolumeUsage_INODES,
			},
		},
	}
	return response, nil
}

// NodeExpandVolume expand node's volume.
func (n nodeServer) NodeExpandVolume(
	ctx context.Context,
	request *csi.NodeExpandVolumeRequest,
) (*csi.NodeExpandVolumeResponse, error) {
	// capacity want to expand
	capacityRange := request.GetCapacityRange()
	// volume path example:
	// /var/lib/kubelet/plugins/kubernetes.io/csi/pv/pvc-4ffaad34-b9f8-4b9f-8a2b-784833796896/globalmount
	volumePath := request.GetVolumePath()
	// persistent volume ID
	volumeId := request.GetVolumeId()

	if capacityRange == nil || capacityRange.RequiredBytes <= 0 {
		msg := "NodeExpandVolume: CapacityRange must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if volumePath == "" {
		msg := "NodeExpandVolume: volumePath must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	// fetch pv info
	pvInfo, err := n.client.CoreV1().PersistentVolumes().Get(ctx, volumeId, metav1.GetOptions{})
	if err != nil {
		msg := fmt.Sprintf("NodeExpandVolume: get pv %s from cluster error: %s", volumeId, err.Error())
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not CSI", log.Any("pv", pvInfo))
		return nil, status.Error(codes.Internal, "pv is not CSI")
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	parameters["volumePath"] = volumePath
	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Errorf("find storage service err:%v", err.Error())
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := storagePlugin.Volumes().NodeExpand(ctx, volumeId, parameters); err != nil {
		log.F(ctx).Errorf("find storage service err:%v", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeExpandVolumeResponse{
		CapacityBytes: capacityRange.RequiredBytes,
	}, nil
}

// return node plugin's support capabilities. such as mount/unmount.
func (n nodeServer) NodeGetCapabilities(
	ctx context.Context,
	request *csi.NodeGetCapabilitiesRequest,
) (*csi.NodeGetCapabilitiesResponse, error) {
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						// capability to mount/unmount volume
						Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						// capability to expand volume
						Type: csi.NodeServiceCapability_RPC_EXPAND_VOLUME,
					},
				},
			},
			{
				Type: &csi.NodeServiceCapability_Rpc{
					Rpc: &csi.NodeServiceCapability_RPC{
						Type: csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
					},
				},
			},
		},
	}, nil
}

// NodeGetInfo return csi node info
// When node-driver-registrar CO plugin register csi driver to kubelet, kubelet will call csi driver's NodeGetInfo
// function.
// If this function success return, it means csi driver has registered.
func (n nodeServer) NodeGetInfo(
	ctx context.Context,
	request *csi.NodeGetInfoRequest,
) (*csi.NodeGetInfoResponse, error) {
	hostname, err := os.Hostname()
	if err != nil {
		log.F(ctx).Errorf("cannot get current hostname:%s", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.NodeGetInfoResponse{
		NodeId: hostname,
	}, nil
}
