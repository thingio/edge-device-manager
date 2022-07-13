package device

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-manager/pkg/utils"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"net/http"
)

const (
	QueryParamProductID     = "product-id"
	QueryParamProductIDDesc = "the identifier of the product"
	QueryParamProductIDType = "string"

	PathParamDeviceID     = "device-id"
	PathParamDeviceIDDesc = "the identifier of the device"
	PathParamDeviceIDType = "string"

	PathParamPropertyID     = "property-id"
	PathParamPropertyIDDesc = "the identifier of the device property, the character '*' represents all properties"
	PathParamPropertyIDType = "string"

	QueryParamPropertyReadType     = "type"
	QueryParamPropertyReadTypeSoft = "soft"
	QueryParamPropertyReadTypeHard = "hard"

	PathParamMethodID     = "method-id"
	PathParamMethodIDDesc = "the identifier of the device method"
	PathParamMethodIDType = "string"

	PathParamEventID     = "event-id"
	PathParamEventIDDesc = "the identifier of the device event"
	PathParamEventIDType = "string"
)

func (r *Resource) createDevice(request *restful.Request, response *restful.Response) {
	device := new(models.Device)
	if err := request.ReadEntity(device); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}
	deviceID := device.ID
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the device's ID is required"))
		return
	}
	productID := device.ProductID
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.Internal.Error("the device[%s]'s ProductID is required", deviceID))
		return
	}

	product, err := r.Manager.GetProduct(productID)
	if err != nil {
		_ = response.WriteError(http.StatusNotFound,
			errors.Internal.Error("the product[%s] is not found", productID))
		return
	}
	_, err = r.Manager.GetDevice(deviceID)
	if err == nil { // verify the duplication of the device
		_ = response.WriteError(http.StatusConflict,
			errors.Internal.Error("the device[%s] is already created", deviceID))
		return
	}

	if err := r.Manager.CreateDevice(product, device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to create the device: %+v", device))
		return
	}
	_ = response.WriteEntity(device)
}
func (r *Resource) deleteDevice(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	if err := r.Manager.DeleteDevice(deviceID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "failed to delete the device[%s]", deviceID))
		return
	}
	_ = response.WriteEntity(struct{}{})
}
func (r *Resource) updateDevice(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	device := new(models.Device)
	if err := request.ReadEntity(device); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("fail to parse the request body"))
	}

	if err := r.Manager.UpdateDevice(device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "failed to update the device[%s]", deviceID))
		return
	}
	_ = response.WriteEntity(device)
}
func (r *Resource) findAllDevices(request *restful.Request, response *restful.Response) {
	productID := request.QueryParameter(QueryParamProductID)
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the query parameter[%s] is required", QueryParamProductID))
		return
	}
	devices, err := r.Manager.ListDevices(productID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	_ = response.WriteEntity(devices)
}
func (r *Resource) findDevice(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	device, err := r.Manager.GetDevice(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	_ = response.WriteEntity(device)
}

func (r *Resource) watchProperties(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	bus, stop, err := r.Manager.SubscribeDeviceProps(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Error("fail to subscribe props for the device[%s]", deviceID))
		return
	}
	if err := utils.SendWSMessage(request, response, bus, stop); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send messages to the device[%s] in a WebSocket connection", deviceID))
	}
}
func (r *Resource) readProperties(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	propertyID := request.PathParameter(PathParamPropertyID)
	if propertyID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamPropertyID))
		return
	}
	readType := request.QueryParameter(QueryParamPropertyReadType)
	if readType == "" {
		readType = QueryParamPropertyReadTypeSoft
	}

	switch readType {
	case QueryParamPropertyReadTypeSoft:
		props, err := r.Manager.Read(deviceID, propertyID)
		if err != nil {
			_ = response.WriteError(http.StatusInternalServerError,
				errors.Internal.Cause(err, "fail to read softly the properties[%s] of the device[%s]", propertyID, deviceID))
			return
		}
		_ = response.WriteEntity(props)
	case QueryParamPropertyReadTypeHard:
		props, err := r.Manager.HardRead(deviceID, propertyID)
		if err != nil {
			_ = response.WriteError(http.StatusInternalServerError,
				errors.Internal.Cause(err, "fail to read hardly the properties[%s] of the device[%s]", propertyID, deviceID))
			return
		}
		_ = response.WriteEntity(props)
	default:
		_ = response.WriteError(http.StatusBadRequest, errors.BadRequest.Error("unsupported read type: %s", readType))
	}
}
func (r *Resource) writeProperties(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	propertyID := request.PathParameter(PathParamPropertyID)
	if propertyID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamPropertyID))
		return
	}
	props := make(map[models.ProductPropertyID]*models.DeviceData)
	if err := request.ReadEntity(&props); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}

	if err := r.Manager.Write(deviceID, propertyID, props); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to write the properties[%s] of the device[%s]", propertyID, deviceID))
		return
	}
	_ = response.WriteEntity(struct{}{})
}
func (r *Resource) callMethod(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	methodID := request.PathParameter(PathParamMethodID)
	if methodID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamMethodID))
		return
	}
	ins := make(map[models.ProductPropertyID]*models.DeviceData)
	if err := request.ReadEntity(&ins); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}

	outs, err := r.Manager.Call(deviceID, methodID, ins)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to call the method[%s] of the device[%s]", methodID, deviceID))
		return
	}
	_ = response.WriteEntity(outs)
}
func (r *Resource) subscribeEvent(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	eventID := request.PathParameter(PathParamEventID)
	if eventID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamEventID))
		return
	}

	bus, stop, err := r.Manager.SubscribeDeviceEvent(deviceID, eventID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to subscribe the event[%s] of the device[%s]", eventID, deviceID))
		return
	}
	if err := utils.SendWSMessage(request, response, bus, stop); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send messages to the device[%s] in a WebSocket connection", deviceID))
		return
	}
}

func (r *Resource) getDevicePropertiesHistory(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	req := new(query.Request)
	if err := request.ReadEntity(req); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}
	data, err := r.Manager.GetDevicePropertiesHistory(deviceID, req)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to get device properties history by request: %+v", req))
	}
	_ = response.WriteEntity(data)
}

func (r *Resource) getDeviceEventsHistory(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	req := new(query.Request)
	if err := request.ReadEntity(req); err != nil {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Cause(err, "fail to parse the request body"))
		return
	}
	data, err := r.Manager.GetDeviceEventsHistory(deviceID, req)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to get device events history by request: %+v", req))
	}
	_ = response.WriteEntity(data)
}
