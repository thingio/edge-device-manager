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
				m.logger.Debugf("the device[%s] is disconnected for some exception, try to reconnect it")
				onlineDevices = append(onlineDevices, device)
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
			if err = m.metaStore.UpdateDevice(device); err != nil {
				m.logger.WithError(err).Errorf("fail to update the device[%s]'s status", device.ID)
			} else {
				m.logger.Debugf("success to update the device[%s]'s status: (%s: %s)",
					device.ID, status.State, status.StateDetail)
			}
		case <-m.ctx.Done():
			stop()
			return
		}
	}
}

func (m *DeviceManager) unregisterDriver(protocolID string, protocol interface{}) {
	m.logger.Debugf("the protocol[%s] has been disconnected", protocolID)
}
