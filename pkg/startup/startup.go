package startup

import (
	"context"
	"github.com/thingio/edge-device-manager/pkg/manager"
	"github.com/thingio/edge-device-manager/pkg/metastore"
)

func Startup(metaStore metastore.MetaStore) {
	ctx, cancel := context.WithCancel(context.Background())

	dm, err := manager.NewDeviceManager(ctx, cancel, metaStore)
	if err != nil {
		panic(err)
	}
	if err = dm.Initialize(); err != nil {
		panic(err)
	}
	if err = dm.Serve(); err != nil {
		panic(err)
	}
}
