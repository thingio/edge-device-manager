package datastore

import (
	"context"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"testing"
)

func TestTDengineDataStore_Connect(t *testing.T) {
	type fields struct {
		dbOpts *config.DataStoreOptions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"Connect to TDengine",
			fields{
				&config.DataStoreOptions{
					Type:      "tdengine",
					Database:  "test",
					BatchSize: 1000,
					TDengine: &config.TDengineOptions{
						Schema:    "tcp",
						Host:      "127.0.0.1",
						Port:      6030,
						Username:  "root",
						Password:  "taosdata",
						Keep:      7,
						Days:      1,
						Blocks:    3,
						Update:    0,
						Precision: "ns",
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg, err := logger.NewLogger(&config.LogOptions{})
			if err != nil {
				t.Error(err)
				return
			}
			ds, err := NewDataStore(context.Background(), lg, tt.fields.dbOpts)
			defer func() {
				if ds != nil {
					ds.Close()
				}
			}()
			if err != nil {
				t.Error(err)
				return
			}
		})
	}
}
