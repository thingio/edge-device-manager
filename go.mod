module github.com/thingio/edge-device-manager

replace github.com/thingio/edge-device-std v0.2.1 => ../edge-device-std

require (
	github.com/apache/arrow/go/v8 v8.0.0
	github.com/emicklei/go-restful-openapi/v2 v2.8.0
	github.com/emicklei/go-restful/v3 v3.7.3
	github.com/go-openapi/spec v0.20.4
	github.com/gobwas/ws v1.1.0
	github.com/influxdata/influxdb v1.9.7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/taosdata/driver-go/v2 v2.0.1-0.20220523115057-e3107e343c03
	github.com/thingio/edge-device-std v0.2.1
	gopkg.in/yaml.v2 v2.4.0
)

go 1.16
