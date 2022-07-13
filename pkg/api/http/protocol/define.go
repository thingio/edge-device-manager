package protocol

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"net/http"
)

const (
	PathParamProtocolID     = "protocol-id"
	PathParamProtocolIDDesc = "the identifier of the protocol"
	PathParamProtocolIDType = "string"
)

func (r Resource) findAllProtocols(request *restful.Request, response *restful.Response) {
	protocols, err := r.Manager.ListProtocols()
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, fmt.Errorf("fail to list all protocols"))
		return
	}
	_ = response.WriteEntity(protocols)
}

func (r Resource) findProtocol(request *restful.Request, response *restful.Response) {
	protocolID := request.PathParameter(PathParamProtocolID)
	if protocolID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the path parameter[%s] is required", PathParamProtocolID))
		return
	}

	protocol, err := r.Manager.GetProtocol(protocolID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, fmt.Errorf("fail to get the protocol"))
		return
	}
	_ = response.WriteEntity(protocol)
}
