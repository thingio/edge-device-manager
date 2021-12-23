package manager

import (
	"context"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	bus "github.com/thingio/edge-device-std/msgbus"
	"github.com/thingio/edge-device-std/operations"
	"time"
)

func NewDeviceManager(ctx context.Context, cancel context.CancelFunc,
	metaStore metastore.MetaStore) (*DeviceManager, error) {
	m := &DeviceManager{
		metaStore: metaStore,

		ctx:    ctx,
		cancel: cancel,
		logger: logger.NewLogger(),
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
	driverExpiration := time.Duration(m.cfg.CommonOptions.DriverHealthCheckIntervalSecond+1) * time.Second
	protocols := cache.New(driverExpiration, driverExpiration)
	protocols.OnEvicted(m.unregisterDriver)
	m.protocols = protocols

	return nil
}

func (m *DeviceManager) Serve() error {
	go m.monitoringDrivers()

	<-m.ctx.Done()
	return nil
}
