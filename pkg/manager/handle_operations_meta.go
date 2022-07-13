package manager

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/thingio/edge-device-std/models"
	"time"
)

func (m *DeviceManager) monitoringDrivers() {
	bus, stop, err := m.ms.SubscribeDriverStatus()
	if err != nil {
		panic(errors.Wrap(err, "fail to subscribe to the statuses of drivers"))
	}

	for {
		select {
		case data := <-bus:
			status, ok := data.(*models.DriverStatus)
			if !ok {
				m.logger.Errorf("invalid format of driver status")
				break
			}
			protocol := status.Protocol
			if protocol == nil || protocol.ID == "" {
				break
			}
			if status.Hello {
				if err = m.initDriver(protocol.ID); err != nil {
					m.logger.WithError(err).Errorf("fail to initialize the protocol driver[%s]", protocol.ID)
					break
				}
			}
			m.protocols.Set(protocol.ID, protocol,
				time.Duration(status.HealthCheckIntervalSecond+1)*time.Second) // set or reset the cache
			m.logger.Debugf("the protocol driver[%s]'s status now is %s", protocol.ID, status.State)
		case <-m.ctx.Done():
			stop()
			return
		}
	}
}

func (m *DeviceManager) initDriver(protocolID string) error {
	products, err := m.metaStore.ListProducts(protocolID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("fail to get products for the protocol[%s]", protocolID))
	}
	onlineDevices := make([]*models.Device, 0)
	for _, product := range products {
		devices, err := m.metaStore.ListDevices(product.ID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("fail to get devices for the product[%s]", product.ID))
		}
		for _, device := range devices {
			switch device.DeviceStatus {
			case models.DeviceStateReconnecting, models.DeviceStateConnected:
				onlineDevices = append(onlineDevices, device)
			case models.DeviceStateException:
				m.logger.Debugf("the device[%s] is disconnected for some exception, try to reconnect it", device.ID)
				onlineDevices = append(onlineDevices, device)
			}

			if err = m.activateRecorder(product, device); err != nil {
				return errors.Wrap(err, fmt.Sprintf("fail to activate recorder for the device[%s]", device.ID))
			}
		}
	}
	if err = m.mc.InitDriver(protocolID, products, onlineDevices); err != nil {
		return err
	}

	go m.monitoringDevices(protocolID)
	return nil
}

func (m *DeviceManager) monitoringDevices(protocolID string) {
	bus, stop, err := m.ms.SubscribeDeviceStatus(protocolID)
	if err != nil {
		m.logger.WithError(err).Errorf("fail to subscribe to the statuses of devices for the driver[%s]", protocolID)
		return
	}

	for {
		select {
		case data := <-bus:
			// TODO cache devices
			status, ok := data.(*models.DeviceStatus)
			if !ok {
				m.logger.Errorf("invalid format of device status")
				break
			}

			device := status.Device
			if device.DeviceStatus == status.State {
				break
			}
			device.DeviceStatus = status.State
			if err = m.UpdateDevice(device); err != nil {
				m.logger.WithError(err).Errorf("fail to update the device[%s]'s status", device.ID)
			} else {
				m.logger.Debugf("success to update the device[%s]'s status: (%s) -> (%s: %s)",
					device.ID, device.DeviceStatus, status.State, status.StateDetail)
			}
		case <-m.ctx.Done():
			stop()
			return
		}
	}
}

func (m *DeviceManager) unregisterDriver(protocolID string, protocol interface{}) {
	m.logger.Debugf("the protocol driver[%s] has been disconnected", protocolID)
}

func (m *DeviceManager) GetProtocol(protocolID string) (*models.Protocol, error) {
	v, ok := m.protocols.Get(protocolID)
	if !ok {
		return nil, errors.New("the protocol[%s] is not found")
	}
	return v.(*models.Protocol), nil
}
func (m *DeviceManager) ListProtocols() ([]*models.Protocol, error) {
	protocols := make([]*models.Protocol, 0)
	for _, protocol := range m.protocols.Items() {
		protocols = append(protocols, protocol.Object.(*models.Protocol))
	}
	return protocols, nil
}

func (m *DeviceManager) CreateProduct(product *models.Product) error {
	if product.Name == "" {
		product.Name = product.ID
	}

	return m.metaStore.CreateProduct(product)
}
func (m *DeviceManager) DeleteProduct(productID string) error {
	product, err := m.GetProduct(productID)
	if err != nil {
		return err
	}
	if err := m.metaStore.DeleteProduct(productID); err != nil {
		return err
	}
	protocolID := product.Protocol
	if err := m.mc.DeleteProduct(protocolID, productID); err != nil {
		return err
	}

	devices, err := m.ListDevices(productID)
	if err != nil {
		return err
	}
	for _, device := range devices {
		if err = m.DeleteDevice(device.ID); err != nil {
			return err
		}
	}
	return nil
}
func (m *DeviceManager) UpdateProduct(product *models.Product) error {
	if err := m.metaStore.UpdateProduct(product); err != nil {
		return err
	} else if err = m.mc.UpdateProduct(product.Protocol, product); err != nil {
		return err
	}
	return nil
}
func (m *DeviceManager) ListProducts(protocolID string) ([]*models.Product, error) {
	return m.metaStore.ListProducts(protocolID)
}
func (m *DeviceManager) GetProduct(productID string) (*models.Product, error) {
	return m.metaStore.GetProduct(productID)
}

func (m *DeviceManager) CreateDevice(product *models.Product, device *models.Device) error {
	if device.Name == "" {
		device.Name = device.ID
	}
	if device.DeviceStatus == "" {
		device.DeviceStatus = models.DeviceStateDisconnected
	}

	if err := m.metaStore.CreateDevice(device); err != nil {
		return err
	} else if err = m.mc.UpdateDevice(product.ID, device); err != nil {
		return err
	}

	if err := m.activateRecorder(product, device); err != nil {
		return err
	}
	return nil
}
func (m *DeviceManager) DeleteDevice(deviceID string) error {
	protocolID, _, err := m.trace(deviceID)
	if err != nil {
		return err
	}
	if err := m.metaStore.DeleteDevice(deviceID); err != nil {
		return err
	} else if err := m.mc.DeleteDevice(protocolID, deviceID); err != nil {
		return err
	}

	if err := m.deactivateRecorder(deviceID); err != nil {
		return err
	}
	return nil
}
func (m *DeviceManager) UpdateDevice(device *models.Device) error {
	protocolID, _, err := m.trace(device.ID)
	if err != nil {
		return err
	}
	if err := m.metaStore.UpdateDevice(device); err != nil {
		return err
	} else if err := m.mc.UpdateDevice(protocolID, device); err != nil {
		return err
	}

	if device.Recording {
		if _, ok := m.recorders.Get(device.ID); !ok {
			product, err := m.GetProduct(device.ProductID)
			if err != nil {
				return err
			}
			if err = m.activateRecorder(product, device); err != nil {
				return err
			}
		}
	} else {
		if err := m.deactivateRecorder(device.ID); err != nil {
			return err
		}
	}
	return nil
}
func (m *DeviceManager) ListDevices(productID string) ([]*models.Device, error) {
	return m.metaStore.ListDevices(productID)
}
func (m *DeviceManager) GetDevice(deviceID string) (*models.Device, error) {
	return m.metaStore.GetDevice(deviceID)
}

func (m *DeviceManager) trace(deviceID string) (protocolID, productID string, err error) {
	if device, err := m.metaStore.GetDevice(deviceID); err != nil {
		return "", "", err
	} else {
		productID = device.ProductID
	}

	if product, err := m.metaStore.GetProduct(productID); err != nil {
		return "", "", err
	} else {
		protocolID = product.Protocol
	}
	return protocolID, productID, nil
}

func (m *DeviceManager) activateRecorder(product *models.Product, device *models.Device) error {
	if !device.Recording {
		return nil
	}

	recorder, err := NewDeviceRecorder(m.logger, product, device, m.ctx, m.ms, m.dataStore)
	if err != nil {
		return err
	}
	go recorder.Start()
	m.recorders.SetDefault(device.ID, recorder)
	return nil
}

func (m *DeviceManager) deactivateRecorder(deviceID string) error {
	if v, ok := m.recorders.Get(deviceID); ok {
		defer m.recorders.Delete(deviceID)
		
		recorder := v.(*Recorder)
		if err := recorder.Stop(false); err != nil {
			return err
		}
	}
	return nil
}
