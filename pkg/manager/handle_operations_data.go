package manager

import (
	"github.com/thingio/edge-device-std/models"
)

func (m *DeviceManager) Subscribe(topic string) (<-chan interface{}, func(), error) {
	return m.ms.Subscribe(topic)
}

func (m *DeviceManager) Read(deviceID string, propertyID string) (map[models.ProductPropertyID]*models.DeviceData, error) {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return nil, err
	}
	return m.mc.Read(protocolID, productID, deviceID, propertyID)
}
func (m *DeviceManager) HardRead(deviceID string, propertyID string) (
	map[models.ProductPropertyID]*models.DeviceData, error) {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return nil, err
	}
	return m.mc.HardRead(protocolID, productID, deviceID, propertyID)
}
func (m *DeviceManager) Write(deviceID string, propertyID string, props map[models.ProductPropertyID]*models.DeviceData) error {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return err
	}
	return m.mc.Write(protocolID, productID, deviceID, propertyID, props)
}
func (m *DeviceManager) Call(deviceID string, methodID string, ins map[models.ProductPropertyID]*models.DeviceData) (
	map[models.ProductPropertyID]*models.DeviceData, error) {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return nil, err
	}
	return m.mc.Call(protocolID, productID, deviceID, methodID, ins)
}
func (m *DeviceManager) SubscribeDeviceProps(deviceID string) (<-chan interface{}, func(), error) {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return nil, nil, err
	}
	return m.ms.SubscribeDeviceProps(protocolID, productID, deviceID, models.DeviceDataMultiPropsID)
}
func (m *DeviceManager) SubscribeDeviceEvent(deviceID string, eventID string) (<-chan interface{}, func(), error) {
	protocolID, productID, err := m.trace(deviceID)
	if err != nil {
		return nil, nil, err
	}
	return m.ms.SubscribeDeviceEvent(protocolID, productID, deviceID, eventID)
}
