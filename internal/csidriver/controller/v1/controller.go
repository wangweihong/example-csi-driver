package v1

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/wangweihong/example-csi-driver/internal/csidriver/paramkey"
	"github.com/wangweihong/example-csi-driver/internal/csidriver/storage"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/wangweihong/eazycloud/pkg/util/maputil"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/wangweihong/eazycloud/pkg/log"

	snapshotv1beta1 "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned/typed/volumesnapshot/v1beta1"
)

var _ csi.ControllerServer = (*controllerServer)(nil)

type controllerServer struct {
	caps           []*csi.ControllerServiceCapability
	nodeID         string
	client         kubernetes.Interface
	snapshotClient *snapshotv1beta1.SnapshotV1beta1Client
}

func NewControllerServer(
	nodeID string,
	clientSet *kubernetes.Clientset,
	snapClient *snapshotv1beta1.SnapshotV1beta1Client,
) *controllerServer {
	return &controllerServer{
		caps: getControllerServiceCapabilities(
			[]csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
				csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
				csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
				csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
				csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
				csi.ControllerServiceCapability_RPC_LIST_VOLUMES_PUBLISHED_NODES,
				csi.ControllerServiceCapability_RPC_GET_CAPACITY,
			}),
		nodeID:         nodeID,
		client:         clientSet,
		snapshotClient: snapClient,
	}
}

func (c *controllerServer) CreateVolume(
	ctx context.Context,
	request *csi.CreateVolumeRequest,
) (*csi.CreateVolumeResponse, error) {
	if request.Name == "" {
		log.F(ctx).Error("request name missing")
		return nil, status.Error(codes.InvalidArgument, "request name missing")
	}

	if request.GetCapacityRange() == nil || request.GetCapacityRange().RequiredBytes <= 0 {
		msg := "CreateVolume CapacityRange must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if request.GetParameters() == nil {
		msg := "CreateVolume Parameters must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	// check pv is created or not.
	parameters := maputil.StringStringMap(request.GetParameters()).DeepCopy()
	pvUID := request.Name // pv
	size := request.GetCapacityRange().RequiredBytes

	volume, exist := getCSIVolume(pvUID)
	if exist {
		log.F(ctx).Warnf("csi volume has created", log.String("volumeName", request.Name))
		return &csi.CreateVolumeResponse{Volume: volume}, nil
	}

	// check if mount_perm is ok
	if mountPermString, ok := parameters["mount_perm"]; ok {
		if _, err := strconv.ParseInt(mountPermString, 0, 32); err != nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"mount_perm is not current file mode:"+mountPermString,
			)
		}
	}

	var sourceSnapshotId string
	// clone/restore from snapshot
	contentSource := request.GetVolumeContentSource()
	if contentSource != nil {
		// create  volume from snapshot
		if contentSnapshot := contentSource.GetSnapshot(); contentSnapshot != nil {
			sourceSnapshotId = contentSnapshot.GetSnapshotId()
			// create volume from volume
		} else if contentVolume := contentSource.GetVolume(); contentVolume != nil {
			log.F(ctx).Errorf("The source %s from volume is not supported", contentSource)
			return nil, status.Error(codes.InvalidArgument, "no source ID provided is invalid")
		} else {
			log.F(ctx).Errorf("The source %s is not snapshot either volume", contentSource)
			return nil, status.Error(codes.InvalidArgument, "no source ID provided is invalid")
		}
	}

	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Error("select storage service from parameter error", log.Any("parameters", parameters), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	csiVolumeID, err := storagePlugin.Volumes().Create(ctx, pvUID, size, sourceSnapshotId, parameters)
	if err != nil {
		log.F(ctx).Error("create volume fail", log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// record backend volume volume in parameters for tracing
	// backend volume id
	parameters[paramkey.BackendStorageVolumeID] = csiVolumeID
	// backend volume source snapshotID
	parameters[paramkey.BackendStorageVolumeSnapshotID] = sourceSnapshotId
	// backend volume name
	parameters[paramkey.BackendStorageVolumeName] = pvUID
	// backend volume name
	parameters[paramkey.BackendStorageVolumeSize] = strconv.FormatInt(request.CapacityRange.RequiredBytes, 10)

	volume = &csi.Volume{
		VolumeId:      pvUID, // pv's name
		CapacityBytes: request.CapacityRange.RequiredBytes,
		VolumeContext: parameters,
	}

	// 如果是基于快照, 必须要返回ContentSource。否则csi-Provisioner会报volume content source missing
	if contentSource != nil {
		volume.ContentSource = request.GetVolumeContentSource()
	}

	setCSIVolume(pvUID, volume)

	return &csi.CreateVolumeResponse{
		Volume: volume,
	}, nil
}

// DeleteVolume delete backend volume
// triggered by pvc delete. when pvc is deleted, csi-privisioner call this function for deleting backend volume
// before delete pv.
// Note that if this function fail, pvc also delete, pv keep in 'Released' state. Pv in this state
// can delete by kubectl command (It is no longer prohibited by pvc controller-manager).
func (c *controllerServer) DeleteVolume(
	ctx context.Context,
	request *csi.DeleteVolumeRequest,
) (*csi.DeleteVolumeResponse, error) {
	if request == nil {
		msg := "request missing"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	pvInfo, err := c.client.CoreV1().
		PersistentVolumes().
		Get(ctx, request.VolumeId, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).Error("get pv info error", log.String("volumeID", request.VolumeId), log.Err(err))
		return nil, status.Error(
			codes.FailedPrecondition,
			fmt.Sprintf("Get Volume: %s from cluster error: %s", request.VolumeId, err.Error()),
		)
	}
	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not csi", log.String("volumeID", request.VolumeId), log.Any("pv.spec", pvInfo.Spec))
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s is not csi", request.VolumeId))
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Error("select storage service from parameter error", log.Any("parameters", parameters), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if err := storagePlugin.Volumes().Delete(ctx, parameters); err != nil {
		log.F(ctx).Errorf("delete volume fail:%w", err)
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	clearCSIVolume(request.VolumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

func (c *controllerServer) ControllerPublishVolume(
	ctx context.Context,
	request *csi.ControllerPublishVolumeRequest,
) (*csi.ControllerPublishVolumeResponse, error) {
	// Volume attachment will be done at node stage process
	return &csi.ControllerPublishVolumeResponse{}, nil
}

func (c *controllerServer) ControllerUnpublishVolume(
	ctx context.Context,
	request *csi.ControllerUnpublishVolumeRequest,
) (*csi.ControllerUnpublishVolumeResponse, error) {
	// Volume attachment will be done at node stage process
	return &csi.ControllerUnpublishVolumeResponse{}, nil
}

func (c *controllerServer) ValidateVolumeCapabilities(
	ctx context.Context,
	request *csi.ValidateVolumeCapabilitiesRequest,
) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeContext:      request.GetVolumeContext(),
			VolumeCapabilities: request.GetVolumeCapabilities(),
			Parameters:         request.GetParameters(),
		},
	}, nil
}

func (c *controllerServer) ListVolumes(
	ctx context.Context,
	request *csi.ListVolumesRequest,
) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controllerServer) GetCapacity(
	ctx context.Context,
	request *csi.GetCapacityRequest,
) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controllerServer) ControllerGetCapabilities(
	ctx context.Context,
	request *csi.ControllerGetCapabilitiesRequest,
) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: c.caps,
	}, nil
}

// CreateSnapshot uses tar command to create snapshot for hostpath volume. The tar command can quickly create
// archives of entire directories. The host image must have "tar" binaries in /bin, /usr/sbin, or /usr/bin.
func (c *controllerServer) CreateSnapshot(
	ctx context.Context,
	request *csi.CreateSnapshotRequest,
) (*csi.CreateSnapshotResponse, error) {
	if strings.TrimSpace(request.SourceVolumeId) == "" {
		log.F(ctx).Error("missing source volume ID")
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	snapshotName := request.GetName()
	if snapshotName == "" {
		log.F(ctx).Error("missing volume snapshot name")
		return nil, status.Error(codes.InvalidArgument, "Snapshot Name missing in request")
	}

	pvInfo, err := c.client.CoreV1().
		PersistentVolumes().
		Get(ctx, request.SourceVolumeId, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).Error("get pv info error", log.String("volumeID", request.SourceVolumeId), log.Err(err))
		return nil, status.Error(
			codes.FailedPrecondition,
			fmt.Sprintf("Get Volume: %s from cluster error: %s", request.SourceVolumeId, err.Error()),
		)
	}
	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not csi", log.String("volumeID", request.SourceVolumeId), log.Any("pv", pvInfo))
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s is not csi", request.SourceVolumeId))
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Error("select storage service from parameter error", log.Any("parameters", parameters), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	snapshot, err := storagePlugin.Snapshots().Create(ctx, request.SourceVolumeId, snapshotName, parameters)
	if err != nil {
		log.F(ctx).Error("create snapshot error", log.String("snapshotName", snapshotName), log.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &csi.CreateSnapshotResponse{
		Snapshot: &csi.Snapshot{
			SizeBytes: snapshot["snapshotSize"].(int64),
			// this id will save in volumeSnapshotContent as name
			SnapshotId:     snapshot["snapshotID"].(string),
			SourceVolumeId: request.SourceVolumeId,
			CreationTime:   &timestamp.Timestamp{Seconds: snapshot["snapshotCreateTime"].(int64)},
			ReadyToUse:     true,
		},
	}, nil
}

func (c *controllerServer) DeleteSnapshot(
	ctx context.Context,
	request *csi.DeleteSnapshotRequest,
) (*csi.DeleteSnapshotResponse, error) {
	// snapshotter生成的快照名为snapshot-<uid>, 而snapshotter生成的volumeSnapshotContent名为snapcontent-<uid>
	volumeSnapInfo, err := storage.ConvertVolumeSnapshotInfoFromSnapshotHandle(request.GetSnapshotId())
	if err != nil {
		msg := "invalid volume snapshotID"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}
	// may be kubernetes bug? volumeSnapshotContent create with prefix "snapcontent", but in volumeSnapshot status
	// reference with prefix "snapshot-"
	volumeSnapshotContentName := strings.ReplaceAll(volumeSnapInfo.Name, "snapshot-", "snapcontent-")
	vsc, err := c.snapshotClient.VolumeSnapshotContents().Get(ctx, volumeSnapshotContentName, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).
			Error("get volume snapshot content error", log.String("snapshotName", volumeSnapshotContentName), log.Err(err))
		// if VolumeSnapshotContents is missing, so backend snapshot missing too, we cannot do anything
		if strings.Contains(err.Error(), "not found") {
			log.F(ctx).Info("delete snapshot success for volumeSnapshotContent missing")
			return &csi.DeleteSnapshotResponse{}, nil
		}

		return nil, status.Error(
			codes.FailedPrecondition,
			fmt.Sprintf(
				"Get Volume Snapshot Content: %s from cluster error: %s",
				volumeSnapshotContentName,
				err.Error(),
			),
		)
	}

	if vsc.Spec.Source.VolumeHandle == nil {
		log.F(ctx).
			Error("volumeSnapshotContent.Spec.Source.VolumeHandle is empty", log.Any("volumeSnapshotContent", vsc))
		return nil, status.Error(codes.FailedPrecondition, "volumeSnapshotContent.Spec.Source.VolumeHandle is empty")
	}

	sourceVolumeID := *vsc.Spec.Source.VolumeHandle
	pvInfo, err := c.client.CoreV1().PersistentVolumes().Get(ctx, sourceVolumeID, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).Error("get persistentVolume error", log.String("pvName", sourceVolumeID), log.Err(err))
		// if pv is missing, so backend snapshot missing too, we cannot do anything
		// return success to avoid endless retry.
		if strings.Contains(err.Error(), "not found") {
			log.F(ctx).Info("delete snapshot success for pv missing")
			return &csi.DeleteSnapshotResponse{}, nil
		}

		return nil, status.Error(
			codes.FailedPrecondition,
			fmt.Sprintf(
				"Get Volume Snapshot Content: %s from cluster error: %s",
				volumeSnapshotContentName,
				err.Error(),
			),
		)
	}
	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not csi", log.String("volumeID", sourceVolumeID), log.Any("pv", pvInfo))
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s is not csi", sourceVolumeID))
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	storagePlugin, err := storage.SelectFactory(ctx, parameters, request.Secrets)
	if err != nil {
		log.F(ctx).Error("select storage service from parameter error", log.Any("parameters", parameters), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if err := storagePlugin.Snapshots().Delete(ctx, volumeSnapInfo.ID, parameters); err != nil {
		log.F(ctx).Error("delete snapshot error", log.String("snaptshotID", volumeSnapInfo.ID), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &csi.DeleteSnapshotResponse{}, nil
}

func (c *controllerServer) ListSnapshots(
	ctx context.Context,
	request *csi.ListSnapshotsRequest,
) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func (c *controllerServer) ControllerExpandVolume(
	ctx context.Context,
	request *csi.ControllerExpandVolumeRequest,
) (*csi.ControllerExpandVolumeResponse, error) {
	capacityRange := request.GetCapacityRange()
	volumeID := request.GetVolumeId()

	if strings.TrimSpace(volumeID) == "" {
		msg := "VolumeID must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	if capacityRange == nil || capacityRange.RequiredBytes <= 0 {
		msg := "CapacityRange must be provided"
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.InvalidArgument, msg)
	}

	pvInfo, err := c.client.CoreV1().
		PersistentVolumes().
		Get(ctx, request.VolumeId, metav1.GetOptions{})
	if err != nil {
		log.F(ctx).Error("get pv info error", log.String("volumeID", volumeID), log.Err(err))
		return nil, status.Error(
			codes.FailedPrecondition,
			fmt.Sprintf("Get Volume: %s from cluster error: %s", volumeID, err.Error()),
		)
	}
	if pvInfo.Spec.CSI == nil {
		log.F(ctx).Error("pv is not csi", log.String("volumeID", volumeID), log.Any("pv", pvInfo))
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("volume %s is not csi", volumeID))
	}

	parameters := maputil.StringStringMap(pvInfo.Spec.CSI.VolumeAttributes).DeepCopy()
	storagePlugin, err := storage.SelectFactory(ctx, parameters, nil)
	if err != nil {
		log.F(ctx).Error("select storage service from parameter error", log.Any("parameters", parameters), log.Err(err))
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	nodeExpansionRequired, err := storagePlugin.Volumes().Expand(ctx, volumeID, parameters, capacityRange.RequiredBytes)
	if err != nil {
		msg := "ControllerExpandVolume Expand TopSC Volume Fail:" + err.Error()
		log.F(ctx).Error(msg)
		return nil, status.Error(codes.FailedPrecondition, msg)
	}

	return &csi.ControllerExpandVolumeResponse{
		// must return capacity
		CapacityBytes: capacityRange.RequiredBytes,
		// tell whether need to expand volume has mounted on node
		// storage type like nfs don't required expand node part. pod will see volume has changed.
		// iscsi required.
		NodeExpansionRequired: nodeExpansionRequired,
	}, nil
}

func (c *controllerServer) ControllerGetVolume(
	ctx context.Context,
	request *csi.ControllerGetVolumeRequest,
) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}

func getControllerServiceCapabilities(
	cl []csi.ControllerServiceCapability_RPC_Type,
) []*csi.ControllerServiceCapability {
	csc := make([]*csi.ControllerServiceCapability, 0, len(cl))

	for _, cap := range cl {
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: cap,
				},
			},
		})
	}

	return csc
}

// record volume has been create success.
var (
	pvcProcessSuccessLock sync.RWMutex
	// csi-provisioner发送创建卷请求后,异常退出.而csi-driver正常执行请求创建csi卷.因此pv没有创建或者没有绑定csi
	// csi-provisioner重启后, 发现pvc没有绑定pvc, 又会发起新的创建请求。导致原来创建的csi卷成为孤儿卷
	// 因此建立一个表, 用于减少孤儿csi卷的创建.
	pvcProcessSuccessMap = map[string]*csi.Volume{}
)

func getCSIVolume(pvUID string) (*csi.Volume, bool) {
	pvcProcessSuccessLock.RLock()
	defer pvcProcessSuccessLock.RUnlock()

	volume, exist := pvcProcessSuccessMap[pvUID]
	return volume, exist
}

func setCSIVolume(pvUID string, volume *csi.Volume) {
	pvcProcessSuccessLock.Lock()
	defer pvcProcessSuccessLock.Unlock()

	pvcProcessSuccessMap[pvUID] = volume
}

func clearCSIVolume(pvUID string) {
	pvcProcessSuccessLock.Lock()
	defer pvcProcessSuccessLock.Unlock()

	delete(pvcProcessSuccessMap, pvUID)
}
