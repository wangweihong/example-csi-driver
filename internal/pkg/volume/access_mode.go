package volume

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	AccessModeReadWrite = "ReadWrite"
	AccessModeReadOnly  = "ReadOnly"
)

func ConvertCSIAccessMode(accessMode csi.VolumeCapability_AccessMode_Mode) string {
	switch accessMode {
	case csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER:
		return AccessModeReadWrite
	case csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY:
		return AccessModeReadOnly
	case csi.VolumeCapability_AccessMode_SINGLE_NODE_SINGLE_WRITER:
		return AccessModeReadWrite
	case csi.VolumeCapability_AccessMode_SINGLE_NODE_MULTI_WRITER:
		return AccessModeReadWrite
	case csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY:
		return AccessModeReadOnly
	case csi.VolumeCapability_AccessMode_MULTI_NODE_SINGLE_WRITER:
		return AccessModeReadWrite
	case csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER:
		return AccessModeReadWrite
	case csi.VolumeCapability_AccessMode_UNKNOWN:
		return ""
	default:
		return ""
	}
}
