package manager

import (
	"context"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/thingio/edge-device-manager/pkg/datastore"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	bus "github.com/thingio/edge-device-std/msgbus"
	"github.com/thingio/edge-device-std/operations"
	"io"
	"net/http"
	"os"
	"time"
)

func NewDeviceManager(ctx context.Context, cancel context.CancelFunc) (*DeviceManager, error) {
	m := &DeviceManager{
		ctx:    ctx,
		cancel: cancel,
	}

	return m, nil
}

type DeviceManager struct {
	// caches
	protocols *cache.Cache
	recorders *cache.Cache

	// operation clients
	mc        operations.ManagerClient
	ms        operations.ManagerService
	metaStore metastore.MetaStore
	dataStore datastore.DataStore

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

	if err := m.initializeStores(); err != nil {
		return err
	}
	if err := m.initializeOperations(); err != nil {
		return err
	}
	if err := m.initializeCaches(); err != nil {
		return err
	}
	return nil
}

func (m *DeviceManager) initializeStores() error {
	metaStore, err := metastore.NewMetaStore(m.ctx, m.logger, m.cfg.ManagerOptions.MetaStoreOptions)
	if err != nil {
		return err
	}
	m.metaStore = metaStore

	dataStore, err := datastore.NewDataStore(m.ctx, m.logger, m.cfg.ManagerOptions.DataStoreOptions)
	if err != nil {
		return err
	}
	m.dataStore = dataStore

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

	recorderExpiration := cache.NoExpiration
	m.recorders = cache.New(recorderExpiration, recorderExpiration)
	return nil
}

func (m *DeviceManager) serve() chan error {
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

func (m *DeviceManager) GetDevicePropertiesHistory(deviceID string, req *query.Request) (*query.Response, error) {
	// FIXME
	return m.dataStore.Query(context.Background(), req)
}
func (m *DeviceManager) GetDeviceEventsHistory(deviceID string, req *query.Request) (*query.Response, error) {
	// FIXME
	return m.dataStore.Query(context.Background(), req)
}
func (m *DeviceManager) ExportDBData(req *query.Request, writer io.Writer, format string) error {
	return m.dataStore.Export(context.Background(), req, writer, format)
}
