package metastore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	productsPath = "products"
	devicesPath  = "devices"

	fileMode os.FileMode = 0664 // not 0x664
)

func NewFileMetaStore(ctx context.Context, lg *logger.Logger, opts *config.FileOptions) (MetaStore, error) {
	root := opts.Path
	if _, err := os.Open(root); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(root, fileMode); err != nil {
				return nil, fmt.Errorf("try to create meta store %s, got %s", root, err.Error())
			}
		} else {
			return nil, fmt.Errorf("invalid path: %s, because %s", root, err.Error())
		}
	}
	return &fileMetaStore{
		root: root,
	}, nil
}

type fileMetaStore struct {
	root string
}

func (s *fileMetaStore) ListProducts(protocolID string) ([]*models.Product, error) {
	root := filepath.Join(s.root, productsPath)
	return s.loadProducts(root, protocolID)
}
func (s *fileMetaStore) loadProducts(root string, protocolID string) ([]*models.Product, error) {
	products := make([]*models.Product, 0)
	if err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		var product *models.Product
		product, err = loadProduct(path)
		if err != nil {
			return err
		}
		if product.Protocol != protocolID {
			return nil
		}
		products = append(products, product)
		return nil
	}); err != nil {
		return nil, err
	}
	return products, nil
}
func (s *fileMetaStore) GetProduct(productID string) (*models.Product, error) {
	path := filepath.Join(s.root, productsPath, fmt.Sprintf("%s.json", productID))
	return loadProduct(path)
}
func loadProduct(path string) (*models.Product, error) {
	product := new(models.Product)
	if err := load(path, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *fileMetaStore) CreateProduct(product *models.Product) error {
	path := filepath.Join(s.root, productsPath, fmt.Sprintf("%s.json", product.ID))
	return save(path, product)
}

func (s *fileMetaStore) DeleteProduct(productID string) error {
	path := filepath.Join(s.root, productsPath, fmt.Sprintf("%s.json", productID))
	return remove(path)
}

func (s *fileMetaStore) UpdateProduct(product *models.Product) error {
	path := filepath.Join(s.root, productsPath, fmt.Sprintf("%s.json", product.ID))
	return update(path, product)
}

func (s *fileMetaStore) ListDevices(productID string) ([]*models.Device, error) {
	path := filepath.Join(s.root, devicesPath)
	return s.loadDevices(path, productID)
}
func (s *fileMetaStore) loadDevices(root string, productID string) ([]*models.Device, error) {
	devices := make([]*models.Device, 0)
	if err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		var device *models.Device
		device, err = loadDevice(path)
		if err != nil {
			return err
		}
		if device.ProductID != productID {
			return nil
		}
		devices = append(devices, device)
		return nil
	}); err != nil {
		return nil, err
	}
	return devices, nil
}
func (s *fileMetaStore) GetDevice(deviceID string) (*models.Device, error) {
	path := filepath.Join(s.root, devicesPath, fmt.Sprintf("%s.json", deviceID))
	return loadDevice(path)
}
func loadDevice(path string) (*models.Device, error) {
	device := new(models.Device)
	if err := load(path, device); err != nil {
		return nil, err
	}
	return device, nil
}

func (s *fileMetaStore) CreateDevice(device *models.Device) error {
	path := filepath.Join(s.root, devicesPath, fmt.Sprintf("%s.json", device.ID))
	return save(path, device)
}

func (s *fileMetaStore) UpdateDevice(device *models.Device) error {
	path := filepath.Join(s.root, devicesPath, fmt.Sprintf("%s.json", device.ID))
	return update(path, device)
}

func (s *fileMetaStore) DeleteDevice(deviceID string) error {
	path := filepath.Join(s.root, devicesPath, fmt.Sprintf("%s.json", deviceID))
	return remove(path)
}

func save(path string, meta interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("fail to marshal the meta configuration, got %s", err.Error())
	}
	return ioutil.WriteFile(path, data, fileMode)
}

func remove(path string) error {
	return os.Remove(path)
}

func update(path string, meta interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, fileMode)
}

func load(path string, meta interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("fail to load the meta configurtion stored in %s, got %s",
			path, err.Error())
	}

	var unmarshaller func([]byte, interface{}) error
	switch ext := filepath.Ext(path); ext {
	case ".json":
		unmarshaller = json.Unmarshal
	case ".yaml", ".yml":
		unmarshaller = yaml.Unmarshal
	default:
		return fmt.Errorf("invalid meta config extension %s, only supporting: json / yaml / yml", ext)
	}

	if err := unmarshaller(data, meta); err != nil {
		return fmt.Errorf("fail to unmarshal the device config, got %s", err.Error())
	}
	return nil
}
