package main

import "github.com/thingio/edge-device-sdk/pkg/models"

func NewFileMetaStore() (*FileMetaStore, error) {
	return &FileMetaStore{}, nil
}

type FileMetaStore struct{}

func (s *FileMetaStore) ListProducts(protocolID string) ([]*models.Product, error) {
	return models.LoadProducts(protocolID), nil
}

func (s *FileMetaStore) ListDevices(productID string) ([]*models.Device, error) {
	return models.LoadDevices(productID), nil
}
