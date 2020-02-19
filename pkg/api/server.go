package api

import (
	"fmt"
	"net/http"

	restful "github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
)

const (
	defaultHttpServerPort    = 8800
	defaultHttpServerAddress = "0.0.0.0"
)

func StartHttpServer() {
	logrus.Infof("Start http server, listening on %d", defaultHttpServerPort)
	listen := fmt.Sprintf("%s:%d", defaultHttpServerAddress, defaultHttpServerPort)

	container := restful.NewContainer()
	container.Add(makeVolumesWebService())
	container.Add(makeHealthWebService())

	http.ListenAndServe(listen, container)
}

func makeVolumesWebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/volumes").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_XML, restful.MIME_JSON)

	ws.Route(ws.GET("/{volume-id}").To(getVolume).
		Doc("get a volume").
		Param(ws.PathParameter("volume-id", "identifier of the volume").DataType("string")).
		Writes(Volume{}).
		Returns(200, "OK", []Volume{}))

	ws.Route(ws.GET("/").To(getVolumes).
		Doc("get all volumes").
		Writes([]Volume{}).
		Returns(200, "OK", []Volume{}))

	ws.Route(ws.POST("").To(createVolume).
		Doc("create a volume"))

	ws.Route(ws.POST("/{volume-id}/mount").To(mountVolume).
		Doc("mount a volume").
		Param(ws.PathParameter("volume-id", "identifier of the volume").DataType("string")).
		Writes(Volume{}).
		Returns(200, "OK", []Volume{}))

	return ws
}

func makeHealthWebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/health").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_XML, restful.MIME_JSON)

	ws.Route(ws.GET("/").To(probeHealth).
		Doc("probe health").
		Returns(200, "OK", ""))

	return ws
}

func probeHealth(request *restful.Request, response *restful.Response) {}
