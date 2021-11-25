package metastore

import (
	"github.com/thingio/edge-device-std/models"
)

type MetaStore interface {
	ListProducts(protocolID string) ([]*models.Product, error)
	CreateProduct(product *models.Product) error
	DeleteProduct(productID string) error
	UpdateProduct(product *models.Product) error
	GetProduct(productID string) (*models.Product, error)

	ListDevices(productID string) ([]*models.Device, error)
	CreateDevice(device *models.Device) error
	DeleteDevice(deviceID string) error
	UpdateDevice(device *models.Device) error
	GetDevice(deviceID string) (*models.Device, error)
}
