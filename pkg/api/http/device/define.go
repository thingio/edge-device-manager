package device

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/thingio/edge-device-std/errors"
	"github.com/thingio/edge-device-std/models"
	"net/http"
	"os"
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

func (r Resource) createDevice(request *restful.Request, response *restful.Response) {
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
	} else if device.Name == "" {
		device.Name = deviceID
	}

	var protocolID string
	productID := device.ProductID
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.Internal.Error("the device[%s]'s product must be specified", deviceID))
		return
	} else {
		if product, err := r.MetaStore.GetProduct(productID); err != nil {
			_ = response.WriteError(http.StatusNotFound,
				errors.Internal.Error("the product[%s] is not found", productID))
			return
		} else {
			device.ProductName = product.Name
			protocolID = product.Protocol
		}
	}
	if _, err := r.MetaStore.GetDevice(deviceID); err == nil { // verify the duplication of the device
		_ = response.WriteError(http.StatusConflict,
			errors.Internal.Error("the device[%s] is already created", deviceID))
		return
	}

	// set the device's status
	if device.DeviceStatus == "" {
		device.DeviceStatus = models.DeviceStateDisconnected
	}

	if err := r.MetaStore.CreateDevice(device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to create the device[%s]", deviceID))
		return
	} else if err = r.OperationClient.UpdateDevice(protocolID, device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send message about creating device[%s] to the driver[%s]", deviceID, protocolID))
		return
	}
	_ = response.WriteEntity(device)
}
func (r Resource) deleteDevice(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	protocolID, _, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
	}
	if err := r.MetaStore.DeleteDevice(deviceID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to delete the device[%s]", deviceID))
		return
	} else if err := r.OperationClient.DeleteDevice(protocolID, deviceID); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send message about deleting device to the driver[%s]", protocolID))
		return
	}
}
func (r Resource) updateDevice(request *restful.Request, response *restful.Response) {
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

	protocolID, _, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
	}
	if err := r.MetaStore.UpdateDevice(device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to update the device[%s]", deviceID))
		return
	} else if err := r.OperationClient.UpdateDevice(protocolID, device); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send message about updating device to the driver[%s]", protocolID))
		return
	}
}
func (r Resource) findAllDevices(request *restful.Request, response *restful.Response) {
	productID := request.QueryParameter(QueryParamProductID)
	if productID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the query parameter[%s] is required", QueryParamProductID))
		return
	}
	devices, err := r.MetaStore.ListDevices(productID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	_ = response.WriteEntity(devices)
}
func (r Resource) findDevice(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest, fmt.Errorf("the path parameter[%s] is required", PathParamDeviceID))
		return
	}
	device, err := r.MetaStore.GetDevice(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError, err)
		return
	}
	_ = response.WriteEntity(device)
}

func (r Resource) watchProperties(request *restful.Request, response *restful.Response) {
	deviceID := request.PathParameter(PathParamDeviceID)
	if deviceID == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("the path parameter[%s] is required", PathParamDeviceID))
		return
	}

	protocolID, productID, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
		return
	}
	bus, stop, err := r.OperationService.SubscribeDeviceProps(protocolID, productID, deviceID, models.DeviceDataMultiPropsID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to watch the properties of the device[%s]", deviceID))
		return
	}
	if err := sendWSMessage(request, response, bus, stop); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send messages to the device[%s] in a WebSocket connection", deviceID))
	}
}
func (r Resource) readProperties(request *restful.Request, response *restful.Response) {
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

	protocolID, productID, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
		return
	}
	var props map[models.ProductPropertyID]*models.DeviceData
	switch readType {
	case QueryParamPropertyReadTypeSoft:
		if props, err = r.OperationClient.Read(protocolID, productID, deviceID, propertyID); err != nil {
			_ = response.WriteError(http.StatusInternalServerError,
				errors.Internal.Cause(err, "fail to read softly the properties[%s] of the device[%s]", propertyID, deviceID))
			return
		}
	case QueryParamPropertyReadTypeHard:
		if props, err = r.OperationClient.HardRead(protocolID, productID, deviceID, propertyID); err != nil {
			_ = response.WriteError(http.StatusInternalServerError,
				errors.Internal.Cause(err, "fail to read hardly the properties[%s] of the device[%s]", propertyID, deviceID))
			return
		}
	}
	_ = response.WriteEntity(props)
}
func (r Resource) writeProperties(request *restful.Request, response *restful.Response) {
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

	protocolID, productID, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
		return
	}
	if err := r.OperationClient.Write(protocolID, productID, deviceID, propertyID, props); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to write the properties[%s] of the device[%s]", propertyID, deviceID))
		return
	}
}
func (r Resource) callMethod(request *restful.Request, response *restful.Response) {
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

	protocolID, productID, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
		return
	}
	outs, err := r.OperationClient.Call(protocolID, productID, deviceID, methodID, ins)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to call the method[%s] of the device[%s]", methodID, deviceID))
		return
	}
	_ = response.WriteEntity(outs)
}
func (r Resource) subscribeEvent(request *restful.Request, response *restful.Response) {
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

	protocolID, productID, err := r.trace(deviceID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to trace the device[%s]", deviceID))
		return
	}
	bus, stop, err := r.OperationService.SubscribeDeviceEvent(protocolID, productID, deviceID, eventID)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to subscribe the event[%s] of the device[%s]", eventID, deviceID))
		return
	}
	if err := sendWSMessage(request, response, bus, stop); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send messages to the device[%s] in a WebSocket connection", deviceID))
		return
	}
}

// sendWSMessage will upgrade an HTTP connection to a WebSocket connection,
// and then sends message from the bus into this connection until it is closed
func sendWSMessage(request *restful.Request, response *restful.Response, bus <-chan interface{}, stop func()) error {
	conn, _, _, err := ws.UpgradeHTTP(request.Request, response.ResponseWriter)
	if err != nil {
		return errors.Internal.Cause(err, "fail to upgrade HTTP as WebSocket")
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer func() {
			cancel()
			_ = conn.Close()
		}()
		for {
			data, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				if err.Error() == "EOF" { // the connection is already disconnected
					return
				}
				continue
			}
			if string(data) == "stop" {
				return
			}
		}
	}()
	for {
		select {
		case event := <-bus:
			if data, err := json.Marshal(event); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, err.Error())
				break
			} else {
				_ = wsutil.WriteServerMessage(conn, ws.OpText, data)
			}
		case <-ctx.Done():
			stop()
			return nil
		}
	}
}

func (r Resource) trace(deviceID string) (protocolID, productID string, err error) {
	if device, err := r.MetaStore.GetDevice(deviceID); err != nil {
		return "", "", err
	} else {
		productID = device.ProductID
	}

	if product, err := r.MetaStore.GetProduct(productID); err != nil {
		return "", "", err
	} else {
		protocolID = product.Protocol
	}
	return protocolID, productID, nil
}
