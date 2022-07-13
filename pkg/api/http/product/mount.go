package product

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
	ws.Path(root).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"PRODUCT META OPERATION"}

	ws.Route(ws.POST("").To(r.createProduct).
		// docs
		Doc("create a new product").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(models.Product{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Product{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	ws.Route(ws.DELETE(fmt.Sprintf("/{%s}", PathParamProductID)).To(r.deleteProduct).
		// docs
		Doc("delete a product by its ID").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter(PathParamProductID, PathParamProductIDDesc).DataType(PathParamProductIDType)).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusNotFound, http.StatusText(http.StatusNotFound), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	ws.Route(ws.PUT(fmt.Sprintf("/{%s}", PathParamProductID)).To(r.updateProduct).
		// docs
		Doc("update a product by its ID").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter(PathParamProductID, PathParamProductIDDesc).DataType(PathParamProductIDType)).
		Reads(models.Product{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Product{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	ws.Route(ws.GET("/").To(r.findAllProducts).
		// docs
		Doc("get all available products").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter(QueryParamProtocolID, QueryParamProtocolIDDesc).DataType(QueryParamProtocolIDType).Required(true)).
		Writes([]models.Product{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []models.Product{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	ws.Route(ws.GET(fmt.Sprintf("/{%s}", PathParamProductID)).To(r.findProduct).
		// docs
		Doc("get an available product by its ID").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.PathParameter(PathParamProductID, PathParamProductIDDesc).DataType(PathParamProductIDType)).
		Writes(models.Product{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Product{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	return ws
}
