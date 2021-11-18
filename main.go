package main

import (
	startup "github.com/thingio/edge-device-sdk/pkg/startup/device_manager"
)

func main() {
	metaStore, err := NewFileMetaStore()
	if err != nil {
		panic(err)
	}
	startup.Startup(metaStore)
}
