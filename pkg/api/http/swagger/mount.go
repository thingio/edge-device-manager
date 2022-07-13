package swagger

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"net/http"
	"path"
)

type Resource struct{}

func (r Resource) WebService(root string) *restful.WebService {
	ws := new(restful.WebService)
	ws.
		Path(root).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_OCTET)
	ws.Route(ws.GET("/").To(getSwaggerUI))
	ws.Route(ws.GET("/{subpath:*}").To(getSwaggerUIResources))
	ws.Route(ws.GET("/swagger.json").To(getSwaggerJson))

	return ws
}

func swaggerUIRoot() string {
	return "public/swagger-ui"
}

func getSwaggerUI(request *restful.Request, response *restful.Response) {
	indexPath := path.Join(swaggerUIRoot(), "index.html")
	http.ServeFile(response.ResponseWriter, request.Request, indexPath)
}

func getSwaggerUIResources(request *restful.Request, response *restful.Response) {
	subPath := path.Join(swaggerUIRoot(), request.PathParameter("subpath"))
	http.ServeFile(response.ResponseWriter, request.Request, subPath)
}

func getSwaggerJson(request *restful.Request, response *restful.Response) {
	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(),
		PostBuildSwaggerObjectHandler: enrichSwaggerObject,
	}
	swagger := restfulspec.BuildSwagger(config)
	swagger.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Device Manager API DOCS",
			Description: "API Docs for ThingIO - Edge Device Manager",
			Version:     "1.0.0",
		},
	}
	swagger.Schemes = []string{"http", "https"}
	_ = response.WriteEntity(swagger)
}

func enrichSwaggerObject(s *spec.Swagger) {
	s.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Edge Device Manager",
			Description: "Manager for Edge devices",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "thingio",
					Email: "thingio@transwarp.io",
					URL:   "https://github.com/thingio",
				},
			},
			Version: "0.1.0",
		},
	}
}
