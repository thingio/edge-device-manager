package main

import (
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-manager/pkg/startup"
)

func main() {
	metaStore, err := metastore.NewFileMetaStore(metastore.DefaultFileMetaStorePath)
	if err != nil {
		panic(err)
	}
	startup.Startup(metaStore)
}
