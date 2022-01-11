package protocol

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-std/models"
	"net/http"
)

const (
	PathParamProtocolID     = "protocol-id"
	PathParamProtocolIDDesc = "the identifier of the protocol"
	PathParamProtocolIDType = "string"
)

func (r Resource) findAllProtocols(request *restful.Request, response *restful.Response) {
	protocols := make([]*models.Protocol, 0)
	for _, item := range r.ProtocolCache.Items() {
		protocols = append(protocols, item.Object.(*models.Protocol))
	}
	_ = response.WriteEntity(protocols)
}

func (r Resource) findProtocol(request *restful.Request, response *restful.Response) {
	protocolID := request.PathParameter(PathParamProtocolID)
	if protocolID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the path parameter[%s] is required", PathParamProtocolID))
		return
	}

	v, ok := r.ProtocolCache.Get(protocolID)
	if !ok {
		_ = response.WriteError(http.StatusNotFound, fmt.Errorf("the protocol[%s] is not found", protocolID))
		return
	}
	_ = response.WriteEntity(v)
}
