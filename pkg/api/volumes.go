package api

import (
	restful "github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"net/http"
	"storage-manager/pkg/fs"
)

type Volume struct {
	Replicas   int
	Size       int
	TargetPath string
}

func getVolume(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	logrus.Infof("Get volumes %v", id)
}

func getVolumes(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("user-id")
	logrus.Infof("Get volumes %v", id)
}

func createVolume(request *restful.Request, response *restful.Response) {
	volume := new(Volume)
	if err := request.ReadEntity(&volume); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
	logrus.Infof("Get volumes %#v", volume)

	mem := fs.NewMemoryFileSystem("/tmp/aa", true)
	go mem.Create()
}

func mountVolume(request *restful.Request, response *restful.Response) {
	volume := new(Volume)
	if err := request.ReadEntity(&volume); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}

	volumeID := request.PathParameter("user-id")
	logrus.Infof("Get volumes user-id: %v, %#v", volumeID, volume)

	mem := fs.NewMemoryFileSystem(volume.TargetPath, true)
	go mem.Create()
}
