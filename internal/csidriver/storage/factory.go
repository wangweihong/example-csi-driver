package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/wangweihong/eazycloud/pkg/log"
)

type Factory interface {
	Volumes() VolumeAPI
	Snapshots() SnapshotAPI
}

type Builder func(ctx context.Context, parameter map[string]string, secret map[string]string) (Factory, error)

var (
	builders    = map[string]Builder{}
	builderLock sync.Mutex
)

func RegisterFactoryBuilder(name string, inst Builder) {
	builderLock.Lock()
	defer builderLock.Unlock()
	if _, exist := builders[name]; exist {
		panic("duplicate register builder type " + name)
	}
	builders[name] = inst
}

func getBuilder(name string) Builder {
	builderLock.Lock()
	defer builderLock.Unlock()

	if svc, exist := builders[name]; exist {
		return svc
	}

	return nil
}

func SelectFactory(ctx context.Context, parameter, secret map[string]string) (Factory, error) {
	storageType, ok := parameter[ExampleStorageTypeKey]
	if !ok {
		return nil, fmt.Errorf("storage type key not exist in parameters:%v", ExampleStorageTypeKey)
	}

	builder := getBuilder(storageType)
	if builder == nil {
		log.F(ctx).Errorf("factory builder  not register: %v", storageType)
		return nil, fmt.Errorf("factory builder  not register: %v", storageType)
	}

	inst, err := builder(ctx, parameter, secret)
	if err != nil {
		log.F(ctx).Errorf("build factory fail for storage type %v err: %v", storageType, err)
		return nil, err
	}
	return inst, nil
}
