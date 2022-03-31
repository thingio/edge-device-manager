package http

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/patrickmn/go-cache"
	"github.com/thingio/edge-device-manager/pkg/api/http/device"
	"github.com/thingio/edge-device-manager/pkg/api/http/product"
	"github.com/thingio/edge-device-manager/pkg/api/http/protocol"
	"github.com/thingio/edge-device-manager/pkg/api/http/swagger"
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-std/operations"
)

const (
	ApiRoot = "/api/v1"
)

func MountAllModules(protocols *cache.Cache, metaStore metastore.MetaStore,
	mc operations.ManagerClient, ms operations.ManagerService) {
	restful.Add(swagger.Resource{}.WebService("/apidocs"))

	restful.Add(protocol.Resource{ProtocolCache: protocols}.WebService(ApiRoot + "/protocols"))
	restful.Add(product.Resource{ProtocolCache: protocols, MetaStore: metaStore, OperationClient: mc}.WebService(ApiRoot + "/products"))
	restful.Add(device.Resource{MetaStore: metaStore, OperationClient: mc, OperationService: ms}.WebService(ApiRoot + "/devices"))
}
