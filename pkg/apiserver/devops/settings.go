package devops

import (
	"github.com/emicklei/go-restful"
	log "k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

func SetupMailServer(req *restful.Request, resp *restful.Response) {
	var server *devops.EmailServerConfig
	err := req.ReadEntity(&server)
	if err != nil {
		log.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
	}

	res, err := devops.SetMailServer(server)
	if err != nil {
		parseErr(err, resp)
		return
	}
	resp.WriteAsJson(res)
}
