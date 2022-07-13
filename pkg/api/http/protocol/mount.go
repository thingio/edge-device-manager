package protocol

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/manager"
	"github.com/thingio/edge-device-std/models"
	"net/http"
)

type Resource struct {
	Manager *manager.DeviceManager
}

func (r Resource) WebService(root string) *restful.WebService {
	ws := new(restful.WebService)
	ws.
		Path(root).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"PROTOCOL META OPERATION"}

	ws.Route(ws.GET("/").To(r.findAllProtocols).
		// docs
		Doc("get all available protocols").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]models.Protocol{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []models.Protocol{}))

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", PathParamProtocolID)).To(r.findProtocol).
		// docs
		Doc("get an available protocol by its ID").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter(PathParamProtocolID, PathParamProtocolIDDesc).DataType(PathParamProtocolIDType)).
		Writes(models.Protocol{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Protocol{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil))

	return ws
}
