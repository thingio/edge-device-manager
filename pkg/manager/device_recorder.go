package manager

import (
	"context"
	"fmt"
	"github.com/thingio/edge-device-manager/pkg/datastore"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	"github.com/thingio/edge-device-std/operations"
)

func NewDeviceRecorder(lg *logger.Logger, product *models.Product, device *models.Device,
	ctx context.Context, dataWatcher operations.DataManagerService, dataStore datastore.DataStore) (*Recorder, error) {
	if lg == nil {
		return nil, errors.NewCommonEdgeErrorWrapper(fmt.Errorf("the logger cannot be nil"))
	}
	if device == nil {
		return nil, errors.NewCommonEdgeErrorWrapper(fmt.Errorf("the device cannot be nil"))
	}
	recorder := &Recorder{
		lg: lg,

		product: product,
		device:  device,

		dataWatcher: dataWatcher,
		dataStore:   dataStore,
	}
	if err := recorder.initialize(ctx); err != nil {
		return nil, err
	}
	return recorder, nil
}

type Recorder struct {
	lg *logger.Logger

	product *models.Product
	device  *models.Device

	dataWatcher operations.DataManagerService
	dataStore   datastore.DataStore

	ctx    context.Context
	cancel context.CancelFunc
}

func (r *Recorder) initialize(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	return nil
}

func (r *Recorder) Start() {
	bus, unsubscribeDeviceProps, err := r.dataWatcher.SubscribeDeviceProps(r.product.Protocol,
		r.product.ID, r.device.ID, "*")
	if err != nil {
		r.lg.WithError(err).Errorf("fail to start the recorder[%s]", r.device.ID)
		return
	}
	defer func() {
		unsubscribeDeviceProps()
	}()

	if err := r.dataStore.CreateTable(&datastore.DeviceDataSchema{
		ProtocolID: r.product.Protocol,
		ProductID:  r.product.ID,
		DeviceID:   r.device.ID,
		Properties: r.product.Properties,
	}); err != nil {
		r.lg.WithError(err).Errorf("fail to create table for device: %s", r.device.ID)
		return
	}

	for {
		select {
		case props := <-bus:
			data, ok := props.(map[models.ProductPropertyID]*models.DeviceData)
			if !ok {
				r.lg.Errorf("interface{} isn't map[*models.ProductPropertyID]*models.DeviceData")
				break
			}
			_ = r.dataStore.Write(&datastore.DeviceDataRecord{
				ProtocolID: r.product.Protocol,
				ProductID:  r.product.ID,
				DeviceID:   r.device.ID,
				Properties: data,
			})
		case <-r.ctx.Done():
			return
		}
	}
}

func (r *Recorder) Stop(force bool) error {
	defer func() {
		r.cancel()
	}()
	return nil
}
