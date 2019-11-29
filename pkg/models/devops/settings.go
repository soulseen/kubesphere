package devops

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	"net/http"
	"strconv"
	"strings"

	cs "kubesphere.io/kubesphere/pkg/simple/client"
)

// set mail server in jenkins
func SetMailServer(server *EmailServerConfig) (*ExecutesResult, error) {

	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	// script for set up mail server
	setMailServerScript := `
import jenkins.model.*

def env = System.getenv()

def emailFromName = "EMAILFROMNAME"
def emailFromAddr = "EMAILFROMADDR"
def emailFromPass = "EMAILFROMPASS"
def emailSmtpHost = "EMAILSMTPHOST"
def emailSmtpPort = "EMAILSMTPPORT"
def ssl = USESSL

def locationConfig = JenkinsLocationConfiguration.get()
locationConfig.adminAddress = "${emailFromName} <${emailFromAddr}>"
locationConfig.save()

def mailer = Jenkins.instance.getDescriptor("hudson.tasks.Mailer")
mailer.setSmtpAuth(emailFromAddr, emailFromPass)
mailer.setReplyToAddress("no-reply@k8s.kubesphere.io")
mailer.setSmtpHost(emailSmtpHost)
mailer.setUseSsl(ssl)
mailer.setSmtpPort(emailSmtpPort)
mailer.save()`

	// replace parameters
	setMailServerScript = strings.Replace(setMailServerScript, "EMAILFROMNAME", server.Email, 1)
	setMailServerScript = strings.Replace(setMailServerScript, "EMAILFROMADDR", server.FromEmailAddr, 1)
	setMailServerScript = strings.Replace(setMailServerScript, "EMAILFROMPASS", server.Password, 1)
	setMailServerScript = strings.Replace(setMailServerScript, "EMAILSMTPHOST", server.EmailHost, 1)
	setMailServerScript = strings.Replace(setMailServerScript, "EMAILSMTPPORT", strconv.Itoa(server.Port), 1)
	setMailServerScript = strings.Replace(setMailServerScript, "USESSL", strconv.FormatBool(server.SslEnable), 1)

	executesMessage, err := jenkinsClient.ExecutesScript(setMailServerScript, server.Submit)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	isSuccess := func(msg *string) bool {
		if *msg == "" {
			return true
		} else {
			return false
		}
	}

	executesResult := &ExecutesResult{
		Success: isSuccess(executesMessage),
		Message: *executesMessage,
	}

	return executesResult, nil

}
