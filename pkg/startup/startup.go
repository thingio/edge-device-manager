package startup

import (
	"context"
	api "github.com/thingio/edge-device-manager/pkg/api/http"
	"github.com/thingio/edge-device-manager/pkg/manager"
)

func Startup() {
	ctx, cancel := context.WithCancel(context.Background())

	dm, err := manager.NewDeviceManager(ctx, cancel)
	if err != nil {
		panic(err)
	}
	if err = dm.Initialize(); err != nil {
		panic(err)
	}
	api.MountAllModules(dm)
	if err = dm.Serve(); err != nil {
		panic(err)
	}
}
