package http

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/api/http/device"
	"github.com/thingio/edge-device-manager/pkg/api/http/naive"
	"github.com/thingio/edge-device-manager/pkg/api/http/product"
	"github.com/thingio/edge-device-manager/pkg/api/http/protocol"
	"github.com/thingio/edge-device-manager/pkg/api/http/swagger"
	"github.com/thingio/edge-device-manager/pkg/manager"
)

const (
	ApiRoot = "/api/v1"
)

func MountAllModules(manager *manager.DeviceManager) {
	restful.Add(swagger.Resource{}.WebService("/apidocs"))

	restful.Add(protocol.Resource{Manager: manager}.WebService(ApiRoot + "/protocols"))
	restful.Add(product.Resource{Manager: manager}.WebService(ApiRoot + "/products"))
	dr := device.Resource{Manager: manager}
	restful.Add(dr.WebService(ApiRoot + "/devices"))

	restful.Add(naive.Resource{Manager: manager}.WebService(ApiRoot))
}
