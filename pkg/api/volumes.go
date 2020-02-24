package api

import (
	"net/http"
	"os"

	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"

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

func (s *Server) createVolume(request *restful.Request, response *restful.Response) {
	volume := new(Volume)
	if err := request.ReadEntity(&volume); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
	logrus.Infof("Create volumes %#v", volume)

	if s.Memory {
		return
	}
}

func (s *Server) deleteVolume(request *restful.Request, response *restful.Response) {
	volume := new(Volume)
	if err := request.ReadEntity(&volume); err != nil {
		response.WriteErrorString(http.StatusInternalServerError, err.Error())
	}
	logrus.Infof("Delete volumes %#v", volume)

	if s.Memory {
		return
	}
}

func (s *Server) mount(request *restful.Request, response *restful.Response) {
	var err error
	var notMnt bool
	volume := new(Volume)

	defer func() {
		if err != nil {
			logrus.Errorf("Request: %s, mount volume error: %v", request.Request.URL, err)
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
		}
	}()

	if err = request.ReadEntity(&volume); err != nil {
		return
	}

	notMnt, err = mount.New("").IsLikelyNotMountPoint(volume.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(volume.TargetPath, 0750); err != nil {
				err = status.Error(codes.Internal, err.Error())
				return
			}
			notMnt = true
		} else {
			err = status.Error(codes.Internal, err.Error())
			return
		}
	}

	if !notMnt {
		return
	}

	volumeID := request.PathParameter("volume-id")
	logrus.Infof("Get volumes volume-id: %s, %#v", volumeID, volume)

	_, err = os.Stat(volume.TargetPath)
	if err != nil {
		return
	}

	// TODO 文件权限

	if s.Memory {
		mem := fs.NewMemoryFileSystem(volume.TargetPath, false)
		mem.Create()
	}

	response.Write([]byte("OK"))
}

func (s *Server) umount(request *restful.Request, response *restful.Response) {
	var err error
	var notMnt bool
	volume := new(Volume)

	defer func() {
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, err.Error())
			logrus.Errorf("Request: %s, mount volume error: %v", request.Request.URL, err)
		}
	}()

	if err = request.ReadEntity(&volume); err != nil {
		return
	}

	notMnt, err = mount.New("").IsLikelyNotMountPoint(volume.TargetPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = status.Error(codes.NotFound, "Targetpath not found")
		} else {
			err = status.Error(codes.Internal, err.Error())
		}
		return
	}
	if notMnt {
		err = status.Error(codes.NotFound, "Volume not mounted")
		return
	}

	volumeID := request.PathParameter("volume-id")
	logrus.Infof("Get volumes volume-id: %s, %#v", volumeID, volume)

	_, err = os.Stat(volume.TargetPath)
	if err != nil {
		return
	}

	err = mount.CleanupMountPoint(volume.TargetPath, mount.New(""), false)
	if err != nil {
		return
	}
}
