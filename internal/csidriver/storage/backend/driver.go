package iscsiexample

//
//import (
//	"context"
//	"fmt"
//	"os"
//
//	"github.com/wangweihong/example-csi-driver/internal/csidriver/paramkey"
//	"github.com/wangweihong/example-csi-driver/internal/csidriver/storage"
//
//	"github.com/wangweihong/eazycloud/pkg/log"
//
//	"github.com/wangweihong/eazycloud/pkg/util/maputil"
//)
//
//const (
//	StorageTypeExample = "iscsiexample"
//)
//
////const (
////	noInitMsg = "backend is not initialized"
////)
//
//// nolint:gochecknoinits
//func init() {
//	storage.RegisterServices(StorageTypeExample, &example{})
//}
//
//type example struct {
//	initialized bool
//	remote      RemoteService
//}
//
//func (e *example) Init(ctx context.Context, secrets map[string]string, parameters map[string]string) error {
//	e.initialized = true
//	e.remote = NewFakeService()
//	return nil
//}
//
//func (e *example) CreateVolume(
//	ctx context.Context,
//	name string,
//	sizeBytes int64,
//	snapshotID string,
//	parameters map[string]string,
//) (string, error) {
//	return "", fmt.Errorf("implement me")
//}
//
//func (e *example) DeleteVolume(ctx context.Context, parameters map[string]string) error {
//	return fmt.Errorf("implement me")
//}
//
//func (e *example) ExpandVolume(
//	ctx context.Context,
//	name string,
//	parameters map[string]string,
//	size int64,
//) (nodeExpansionRequired bool, err error) {
//	return false, nil
//}
//
//func (e *example) AttachVolume(ctx context.Context, s string, m map[string]interface{}) error {
//	return nil
//}
//
//func (e *example) DetachVolume(ctx context.Context, s string, m map[string]interface{}) error {
//	return nil
//}
//
//func (e *example) StageVolume(
//	ctx context.Context,
//	name string,
//	parameters map[string]string,
//	mountPerm os.FileMode,
//) error {
//	return fmt.Errorf("implement me")
//}
//
//// UnstageVolume umount iscsi volume from remote storage service to  pod running host.
//func (e *example) UnstageVolume(ctx context.Context, name string, parameters map[string]string) error {
//	targetPath := maputil.StringStringMap(parameters).Get(paramkey.BackendStorageVolumeStagePath)
//	if targetPath == "" {
//		msg := "target path is empty"
//		log.F(ctx).Error(msg, log.String("paramKey", paramkey.BackendStorageVolumeStagePath))
//		return fmt.Errorf(msg)
//	}
//	return nil
//}
//
//func (e *example) NodeExpandVolume(ctx context.Context, name string, parameters map[string]string) error {
//	return fmt.Errorf("implement me")
//}
//
//func (e *example) CreateSnapshot(
//	ctx context.Context,
//	s string,
//	s2 string,
//	m map[string]string,
//) (map[string]interface{}, error) {
//	return nil, fmt.Errorf("implement me")
//}
//
//func (e *example) DeleteSnapshot(ctx context.Context, volumeSnapshotID string, parameters map[string]string) error {
//	return fmt.Errorf("implement me")
//}
//
////func (e *example) checkInit(ctx context.Context) error {
////
////	return nil
////}
