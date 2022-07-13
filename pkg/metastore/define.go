package metastore

import (
	"context"
	"fmt"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
)

func NewMetaStore(ctx context.Context, lg *logger.Logger, opts *config.MetaStoreOptions) (MetaStore, error) {
	var ms MetaStore
	switch opts.Type {
	case config.MetaStoreTypeFile:
		if opts.File == nil {
			return nil, fmt.Errorf("the configuration for file system is required")
		}
		fms, err := NewFileMetaStore(ctx, lg, opts.File)
		if err != nil {
			return nil, err
		}
		ms = fms
	default:
		return nil, fmt.Errorf("unsupported datastore type: %s", opts.Type)
	}
	return ms, nil
}

type MetaStore interface {
	ListProducts(protocolID string) ([]*models.Product, error)
	// CreateProduct doesn't verify the duplication of the product, so it's up to the upper business.
	CreateProduct(product *models.Product) error
	DeleteProduct(productID string) error
	UpdateProduct(product *models.Product) error
	GetProduct(productID string) (*models.Product, error)

	ListDevices(productID string) ([]*models.Device, error)
	// CreateDevice doesn't verify the duplication of the device, so it's up to the upper business.
	CreateDevice(device *models.Device) error
	DeleteDevice(deviceID string) error
	UpdateDevice(device *models.Device) error
	GetDevice(deviceID string) (*models.Device, error)
}
