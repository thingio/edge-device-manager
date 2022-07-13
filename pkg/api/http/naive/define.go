package naive

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"github.com/thingio/edge-device-manager/pkg/datastore/query"
	"github.com/thingio/edge-device-manager/pkg/utils"
	"github.com/thingio/edge-device-std/errors"
	"net/http"
	"net/url"
)

const (
	QueryParamTopic     = "topic"
	QueryParamTopicDesc = ""
	QueryParamTopicType = "string"

	QueryParamRawQuery     = "raw-query"
	QueryParamRawQueryDesc = ""
	QueryParamRawQueryType = "string"

	QueryParamDataExportFormat     = "export-format"
	QueryParamDataExportFormatDesc = "the export format of device history data, such as parquet, csv..."
	QueryParamDataExportFormatType = "string"
)

func (r Resource) exportDBData(request *restful.Request, response *restful.Response) {
	format := request.QueryParameter(QueryParamDataExportFormat)
	if format == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("lack of query parameter: %s", QueryParamDataExportFormat))
		return
	}
	raw := request.QueryParameter(QueryParamRawQuery)
	if format == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("lack of query parameter: %s", QueryParamRawQuery))
		return
	}

	req := new(query.Request)
	req.Raw = raw
	fileName := fmt.Sprintf("query.%s", format)
	response.Header().Add("Content-Type", "application/octet-stream")
	response.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, url.QueryEscape(fileName)))
	response.Header().Add("File-Name", url.QueryEscape(fileName))
	if err := r.Manager.ExportDBData(req, response.ResponseWriter, format); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to export data by request: %+v", req))
		return
	}
}

func (r Resource) watchMessageBusTopic(request *restful.Request, response *restful.Response) {
	topic := request.QueryParameter(QueryParamTopic)
	if topic == "" {
		_ = response.WriteError(http.StatusBadRequest,
			errors.BadRequest.Error("lack of query parameter: %s", QueryParamTopic))
		return
	}

	bus, stop, err := r.Manager.Subscribe(topic)
	if err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to subscribe the topic[%s]", topic))
		return
	}
	if err := utils.SendWSMessage(request, response, bus, stop); err != nil {
		_ = response.WriteError(http.StatusInternalServerError,
			errors.Internal.Cause(err, "fail to send messages of the topic[%s] in a WebSocket connection", topic))
		return
	}
}
