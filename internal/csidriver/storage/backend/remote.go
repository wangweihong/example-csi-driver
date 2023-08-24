package iscsiexample

type RemoteService interface {
	CreateVolume(name string) error
	DeleteVolume(name string) error
	ExpandVolume(name string) error
	GetVolumeMountList(name string) error
	CreateSnapshot(volumeID string, snapshotName string) error
	DeleteSnapshot(snapshotID string) error
}

func NewFakeService() RemoteService {
	return &fake{}
}

type fake struct{}

func (fake) CreateVolume(name string) error {
	panic("implement me")
}

func (fake) DeleteVolume(name string) error {
	panic("implement me")
}

func (fake) ExpandVolume(name string) error {
	panic("implement me")
}

func (fake) GetVolumeMountList(name string) error {
	panic("implement me")
}

func (fake) CreateSnapshot(volumeID string, snapshotName string) error {
	panic("implement me")
}

func (fake) DeleteSnapshot(snapshotID string) error {
	panic("implement me")
}
