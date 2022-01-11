package device

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/metastore"
	"github.com/thingio/edge-device-std/models"
	"github.com/thingio/edge-device-std/operations"
	"net/http"
)

type Resource struct {
	MetaStore        metastore.MetaStore
	OperationClient  operations.ManagerClient
	OperationService operations.ManagerService
}

func (r Resource) WebService(root string) *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(root).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	// DEVICE META OPERATIONS

	metaTags := []string{"DEVICE META OPERATION"}

	ws.Route(ws.POST("").To(r.createDevice).
		// docs
		Doc("create a new device").
		Metadata(restfulspec.KeyOpenAPITags, metaTags).
		Reads(models.Device{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Device{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.DELETE(fmt.Sprintf("/{%s}", PathParamDeviceID)).To(r.deleteDevice).
		// docs
		Doc("delete a device by its ID").
		Metadata(restfulspec.KeyOpenAPITags, metaTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.PUT(fmt.Sprintf("/{%s}", PathParamDeviceID)).To(r.updateDevice).
		// docs
		Doc("update a device by its ID").
		Metadata(restfulspec.KeyOpenAPITags, metaTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Reads(models.Device{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Device{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.GET("/").To(r.findAllDevices).
		// docs
		Doc("get all available devices").
		Metadata(restfulspec.KeyOpenAPITags, metaTags).
		Param(ws.QueryParameter(QueryParamProductID, QueryParamProductIDDesc).DataType(QueryParamProductIDType)).
		Writes([]models.Device{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []models.Device{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.GET(fmt.Sprintf("/{%s}", PathParamDeviceID)).To(r.findDevice).
		// docs
		Doc("get an available device by its ID").
		Metadata(restfulspec.KeyOpenAPITags, metaTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Writes(models.Device{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), models.Device{}).
		Returns(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))

	// DEVICE DATA OPERATIONS

	dataTags := []string{"DEVICE DATA OPERATION"}

	ws.Route(ws.GET(fmt.Sprintf("/{%s}/properties", PathParamDeviceID)).To(r.watchProperties).
		// docs
		Doc("watch the device properties").
		Notes("When you want to observe changes of device properties for a long time, it is more efficient than polling to call the read interface"+
			"The connection will upgrade as the websocket protocol.\n"+
			"For example, we can open a websocket connection using 'ws://127.0.0.1:8080/api/v1/devices/randnum_test01/properties';\n"+
			"How to test quickly? \n"+
			"  1. open the browser and visit the webpage 'http://www.jsons.cn/websocket/';\n"+
			"  2. input 'ws://127.0.0.1:10996/api/v1/devices/randnum_test01/properties' to replace '请输入要检测的Websocket地址，例如：ws://localhost:8080/wssajax' and click the button 'Websocket连接';\n"+
			"  3. the textbox will prompt '连接成功，现在你可以发送信息进行测试了！';\n"+
			"  4. the textbox will prompt if the device push an event:\n服务端回应 2022-01-11 17:34:25\n{\"float\":{\"name\":\"float\",\"type\":\"float\",\"value\":0.3038871248932524,\"ts\":\"2022-01-11T17:34:25.031244879+08:00\"},\"int\":{\"name\":\"int\",\"type\":\"int\",\"value\":8,\"ts\":\"2022-01-11T17:34:25.031231535+08:00\"}}\n"+
			"  5. input 'stop' to replace '请输入测试消息', if you want to unsubscribe the event;\n"+
			"  6. the textbox will prompt:\n你发送的信息 2022-01-11 17:34:40\nstop\nWebsocket连接已断开！",
		).
		Metadata(restfulspec.KeyOpenAPITags, dataTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil))
	ws.Route(ws.GET(fmt.Sprintf("/{%s}/properties/{%s}", PathParamDeviceID, PathParamPropertyID)).To(r.readProperties).
		// docs
		Doc("read the device properties").
		Metadata(restfulspec.KeyOpenAPITags, dataTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Param(ws.PathParameter(PathParamPropertyID, PathParamPropertyIDDesc).DataType(PathParamPropertyIDType)).
		Param(ws.QueryParameter(QueryParamPropertyReadType, "read property from the cache(soft) or the real device(hard)").
			DataType("string").
			Required(false).
			PossibleValues([]string{QueryParamPropertyReadTypeSoft, QueryParamPropertyReadTypeHard}).
			DefaultValue(QueryParamPropertyReadTypeSoft)).
		Writes(map[models.ProductPropertyID]models.DeviceData{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), map[models.ProductPropertyID]models.DeviceData{}).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.PUT(fmt.Sprintf("/{%s}/properties/{%s}", PathParamDeviceID, PathParamPropertyID)).To(r.writeProperties).
		// docs
		Doc("write the device properties").
		Metadata(restfulspec.KeyOpenAPITags, dataTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Param(ws.PathParameter(PathParamPropertyID, PathParamPropertyIDDesc).DataType(PathParamPropertyIDType)).
		Reads(map[models.ProductPropertyID]models.DeviceData{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.POST(fmt.Sprintf("/{%s}/methods/{%s}", PathParamDeviceID, PathParamMethodID)).To(r.callMethod).
		// docs
		Doc("call the device method").
		Metadata(restfulspec.KeyOpenAPITags, dataTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Param(ws.PathParameter(PathParamMethodID, PathParamMethodIDDesc).DataType(PathParamMethodIDType)).
		Reads(map[models.ProductPropertyID]models.DeviceData{}).
		Writes(map[models.ProductPropertyID]models.DeviceData{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), map[models.ProductPropertyID]models.DeviceData{}).
		Returns(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), nil))
	ws.Route(ws.GET(fmt.Sprintf("/{%s}/events/{%s}", PathParamDeviceID, PathParamEventID)).To(r.subscribeEvent).
		// docs
		Doc("subscribe the device event").
		Notes("The connection will upgrade as the websocket protocol.\n"+
			"For example, we can open a websocket connection using 'ws://127.0.0.1:8080/api/v1/devices/randnum_test01/events/test';\n"+
			"How to test quickly? \n"+
			"  1. open the browser and visit the webpage 'http://www.jsons.cn/websocket/';\n"+
			"  2. input 'ws://127.0.0.1:10996/api/v1/devices/randnum_test01/events/test' to replace '请输入要检测的Websocket地址，例如：ws://localhost:8080/wssajax' and click the button 'Websocket连接';\n"+
			"  3. the textbox will prompt '连接成功，现在你可以发送信息进行测试了！';\n"+
			"  4. the textbox will prompt if the device push an event:\n服务端回应 2022-01-11 17:34:25\n{\"float\":{\"name\":\"float\",\"type\":\"float\",\"value\":0.3038871248932524,\"ts\":\"2022-01-11T17:34:25.031244879+08:00\"},\"int\":{\"name\":\"int\",\"type\":\"int\",\"value\":8,\"ts\":\"2022-01-11T17:34:25.031231535+08:00\"}}\n"+
			"  5. input 'stop' to replace '请输入测试消息', if you want to unsubscribe the event;\n"+
			"  6. the textbox will prompt:\n你发送的信息 2022-01-11 17:34:40\nstop\nWebsocket连接已断开！",
		).
		Metadata(restfulspec.KeyOpenAPITags, dataTags).
		Param(ws.PathParameter(PathParamDeviceID, PathParamDeviceIDDesc).DataType(PathParamDeviceIDType)).
		Param(ws.PathParameter(PathParamEventID, PathParamEventIDDesc).DataType(PathParamEventIDType)).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil))

	return ws
}
