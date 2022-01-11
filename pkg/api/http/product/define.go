package product

import (
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"net/http"
)

const (
	QueryParamProtocolID     = "protocol-id"
	QueryParamProtocolIDDesc = "the identifier of the protocol"
	QueryParamProtocolIDType = "string"

	PathParamProductID     = "product-id"
	PathParamProductIDDesc = "the identifier of the product"
	PathParamProductIDType = "string"
)

func (r Resource) createProduct(request *restful.Request, response *restful.Response) {
	product := new(models.Product)
	if err := request.ReadEntity(product); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}
	productID := product.ID
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the product's ID is required"))
		return
	} else if product.Name == "" {
		product.Name = productID
	}
	protocolID := product.Protocol
	if protocolID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.Internal.Error("the product[%s]'s protocol must be specified", productID))
		return
	} else {
		_, ok := r.ProtocolCache.Get(protocolID)
		if !ok {
			_ = response.WriteError(http.StatusNotFound,
				errors.Internal.Error("the protocol[%s] is not available yet", protocolID))
			return
		}
	}
	if _, err := r.MetaStore.GetProduct(productID); err == nil { // verify the duplication of the product
		_ = response.WriteError(http.StatusConflict,
			errors.Internal.Error("the product[%s] is already created", productID))
		return
	}

	if err := r.MetaStore.CreateProduct(product); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to create the product[%s]", productID))
		return
	}
	_ = response.WriteEntity(product)
}

func (r Resource) deleteProduct(request *restful.Request, response *restful.Response) {
	productID := request.PathParameter(PathParamProductID)
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamProductID))
		return
	}

	var protocolID string
	if product, err := r.MetaStore.GetProduct(productID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to find the product[%s]", productID))
		return
	} else {
		protocolID = product.Protocol
	}

	if err := r.MetaStore.DeleteProduct(productID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to delete the product[%s]", productID))
		return
	} else {
		devices, err := r.MetaStore.ListDevices(productID)
		if err != nil {
			_ = response.WriteError(http.StatusInternalServerError,
				errors.Internal.Cause(err, "fail to get devices derived from the product[%s]", productID))
			return
		}
		for _, device := range devices {
			_ = r.MetaStore.DeleteDevice(device.ID)
		}
	}
	if err := r.OperationClient.DeleteProduct(protocolID, productID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send message about deleting product to the driver[%s]", protocolID))
		return
	}
}

func (r Resource) updateProduct(request *restful.Request, response *restful.Response) {
	productID := request.PathParameter(PathParamProductID)
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamProductID))
		return
	}
	product := new(models.Product)
	if err := request.ReadEntity(product); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("fail to parse the request body"))
		return
	}
	protocolID := product.Protocol
	if protocolID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the product[%s]'s protocol must be specified", productID))
		return
	}

	if err := r.MetaStore.UpdateProduct(product); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to update the product[%s]", productID))
		return
	} else if err = r.OperationClient.UpdateProduct(protocolID, product); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send message about updating product to the driver[%s]", protocolID))
		return
	}
}

func (r Resource) findAllProducts(request *restful.Request, response *restful.Response) {
	protocolID := request.QueryParameter(QueryParamProtocolID)
	if protocolID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the query parameter[%s] is required", QueryParamProtocolID))
		return
	}
	products, err := r.MetaStore.ListProducts(protocolID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to get products classified as the protocol[%s]", protocolID))
		return
	}
	_ = response.WriteEntity(products)
}

func (r Resource) findProduct(request *restful.Request, response *restful.Response) {
	productID := request.PathParameter(PathParamProductID)
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamProductID))
		return
	}
	product, err := r.MetaStore.GetProduct(productID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	_ = response.WriteEntity(product)
}
