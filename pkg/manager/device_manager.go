package manager

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	api "github.com/thingio/edge-device-manager/pkg/api/http"
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	bus "github.com/thingio/edge-device-std/msgbus"
	"github.com/thingio/edge-device-std/operations"
	"net/http"
	"os"
	"time"
)

func NewDeviceManager(ctx context.Context, cancel context.CancelFunc,
	metaStore metastore.MetaStore) (*DeviceManager, error) {
	m := &DeviceManager{
		metaStore: metaStore,

		ctx:    ctx,
		cancel: cancel,
	}

	return m, nil
}

type DeviceManager struct {
	// caches
	protocols *cache.Cache

	// operation clients
	mc        operations.ManagerClient
	ms        operations.ManagerService
	metaStore metastore.MetaStore

	// lifetime control variables for the device driver
	ctx    context.Context
	cancel context.CancelFunc
	logger *logger.Logger
	cfg    *config.Configuration
}

func (m *DeviceManager) Initialize() error {
	if cfg, err := config.NewConfiguration(); err != nil {
		return err
	} else {
		m.cfg = cfg
	}
	if lg, err := logger.NewLogger(&m.cfg.LogOptions); err != nil {
		return err
	} else {
		m.logger = lg
	}

	if err := m.initializeOperations(); err != nil {
		return err
	}
	if err := m.initializeCaches(); err != nil {
		return err
	}
	return nil
}

func (m *DeviceManager) initializeOperations() error {
	mb, err := bus.NewMessageBus(&m.cfg.MessageBus, m.logger)
	if err != nil {
		return errors.Wrap(err, "fail to initialize the message bus")
	}

	mc, err := operations.NewManagerClient(mb, m.logger)
	if err != nil {
		return errors.Wrap(err, "fail to new an operations client")
	}
	m.mc = mc
	ms, err := operations.NewManagerService(mb, m.logger)
	if err != nil {
		return errors.Wrap(err, "fail to new an operations service")
	}
	m.ms = ms

	return nil
}

func (m *DeviceManager) initializeCaches() error {
	// initialize the drivers cache
	driverExpiration := 1 * time.Minute
	protocols := cache.New(driverExpiration, driverExpiration)
	protocols.OnEvicted(m.unregisterDriver)
	m.protocols = protocols

	return nil
}

func (m *DeviceManager) serve() chan error {
	api.MountAllModules(m.protocols, m.metaStore, m.mc, m.ms)
	errs := make(chan error)
	go func() {
		addr := fmt.Sprintf(":%d", m.cfg.ManagerOptions.HTTP.Port)
		m.logger.Infof("the HTTP server's address is %s", addr)
		if err := http.ListenAndServe(addr, restful.DefaultContainer); err != nil {
			errs <- err
		}
	}()
	return nil
}

func (m *DeviceManager) Serve() error {
	go m.monitoringDrivers()

	errs := m.serve()
	select {
	case err := <-errs:
		panic(err)
	case <-m.ctx.Done():
		os.Exit(1)
	}
	return nil
}
