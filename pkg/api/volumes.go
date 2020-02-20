package api

import (
	restful "github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
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

	mem := fs.NewMemoryFileSystem("/tmp", true)
	go mem.Create()
}

func mountVolume(request *restful.Request, response *restful.Response) {
	var err error
	volume := new(Volume)

	defer func() {
		if err != nil {
			logrus.Errorf("Request: %s, mount volume error: %v", request.Request.URL, err)
		}
	}()

	if err = request.ReadEntity(&volume); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	volumeID := request.PathParameter("user-id")
	logrus.Infof("Get volumes user-id: %s, %#v", volumeID, volume)

	_, err = os.Stat(volume.TargetPath)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO 文件权限

	mem := fs.NewMemoryFileSystem(volume.TargetPath, true)
	go mem.Create()

	response.Write([]byte("OK"))
}
