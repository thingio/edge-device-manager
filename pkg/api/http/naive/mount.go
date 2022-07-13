package naive

import (
	"fmt"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/manager"
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

	tags := []string{"NAIVE DATA OPERATION"}

	ws.Route(ws.GET(fmt.Sprintf("/db/data:export")).To(r.exportDBData).
		Doc("").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter(QueryParamDataExportFormat, QueryParamDataExportFormatDesc).DataType(QueryParamDataExportFormatType).
			Required(true).PossibleValues([]string{"parquet", "csv"})).
		Param(ws.QueryParameter(QueryParamRawQuery, QueryParamRawQueryDesc).DataType(QueryParamRawQueryType).Required(true)).
		Produces("application/octet-stream").
		Writes([]byte{}).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), []byte{}))
	ws.Route(ws.GET(fmt.Sprintf("/ms/topic:watch")).To(r.watchMessageBusTopic).
		// docs
		Doc("watch the device properties using WebSocket").
		Notes("When you want to watch the specific topic from our message bus, you can use this!"+
			"The connection will upgrade as the <b>WebSocket</b> protocol.\n"+
			"For example, we can open a websocket connection using 'ws://127.0.0.1:10996/api/v1/ms/topic:watch?topic=[topic you want to watch]';\n"+
			"The format of topic please refer to 'https://github.com/thingio/edge-device-std/tree/main/docs/zh#topic-%E7%BA%A6%E5%AE%9A' />"+
			"How to test quickly? \n"+
			"  1. open the browser and visit the webpage 'http://www.jsons.cn/websocket/';\n"+
			"  2. input 'ws://127.0.0.1:10996/api/v1/ms/topic:watch?topic=[topic you want to watch]' to replace '请输入要检测的Websocket地址，例如：ws://localhost:8080/wssajax' and click the button 'Websocket连接';\n"+
			"  3. the textbox will prompt '连接成功，现在你可以发送信息进行测试了！';\n"+
			"  4. input 'stop' to replace '请输入测试消息', if you want to unsubscribe the event;\n"+
			"  5. the textbox will prompt:\n你发送的信息 2022-01-11 17:34:40\nstop\nWebsocket连接已断开！",
		).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Param(ws.QueryParameter(QueryParamTopic, QueryParamTopicDesc).DataType(QueryParamTopicType).Required(true)).
		Returns(http.StatusOK, http.StatusText(http.StatusOK), nil))
	return ws
}
