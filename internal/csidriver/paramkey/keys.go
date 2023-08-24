package paramkey

// common key store backend volume info in parameters.
const (
	BackendStoragePrefix = "BackendStorageVolume/"
	// name of storage service.
	BackendStorageVolumeID         = BackendStoragePrefix + "VolumeID"
	BackendStorageVolumeName       = BackendStoragePrefix + "VolumeName"
	BackendStorageVolumeSize       = BackendStoragePrefix + "VolumeSize"
	BackendStorageVolumeSnapshotID = BackendStoragePrefix + "VolumeSnapshotID"
	// which snapshot volume create from.
	BackendStorageVolumeSourceSnapshotID = BackendStoragePrefix + "VolumeSourceSnapshotID"

	// Stage Parameter Keys.
	BackendStorageVolumeStagePath       = BackendStoragePrefix + "StagePath"
	BackendStorageVolumeMountFlag       = BackendStoragePrefix + "MountFlag"
	BackendStorageVolumeFilesystemType  = BackendStoragePrefix + "FsType"
	BackendStorageVolumeMountPermission = BackendStoragePrefix + "MountPerm"
)
