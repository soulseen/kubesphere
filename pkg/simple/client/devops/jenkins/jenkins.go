// Copyright 2015 Vadim Kravcenko
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Gojenkins is a Jenkins Client in Go, that exposes the jenkins REST api in a more developer friendly way.
package jenkins

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

type Jenkins struct {
	Server    string
	Version   string
	Raw       *ExecutorResponse
	Requester *Requester
}

// Loggers
var (
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

// Init Method. Should be called after creating a Jenkins Instance.
// e.g jenkins := CreateJenkins("url").Init()
// HTTP Client is set here, Connection to jenkins is tested here.
func (j *Jenkins) Init() (*Jenkins, error) {
	j.initLoggers()

	// Check Connection
	j.Raw = new(ExecutorResponse)
	rsp, err := j.Requester.GetJSON("/", j.Raw, nil)
	if err != nil {
		return nil, err
	}

	j.Version = rsp.Header.Get("X-Jenkins")
	if j.Raw == nil {
		return nil, errors.New("Connection Failed, Please verify that the host and credentials are correct.")
	}

	return j, nil
}

func (j *Jenkins) initLoggers() {
	Info = log.New(os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(os.Stdout,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(os.Stderr,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// Create a new folder
// This folder can be nested in other parent folders
// Example: jenkins.CreateFolder("newFolder", "grandparentFolder", "parentFolder")
func (j *Jenkins) CreateFolder(name, description string, parents ...string) (*Folder, error) {
	folderObj := &Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, name), "/job/")}
	folder, err := folderObj.Create(name, description)
	if err != nil {
		return nil, err
	}
	return folder, nil
}

// Create a new job in the folder
// Example: jenkins.CreateJobInFolder("<config></config>", "newJobName", "myFolder", "parentFolder")
func (j *Jenkins) CreateJobInFolder(config string, jobName string, parentIDs ...string) (*Job, error) {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, jobName), "/job/")}
	qr := map[string]string{
		"name": jobName,
	}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Create a new job from config File
// Method takes XML string as first parameter, and if the name is not specified in the config file
// takes name as string as second parameter
// e.g jenkins.CreateJob("<config></config>","newJobName")
func (j *Jenkins) CreateJob(config string, options ...interface{}) (*Job, error) {
	qr := make(map[string]string)
	if len(options) > 0 {
		qr["name"] = options[0].(string)
	} else {
		return nil, errors.New("Error Creating Job, job name is missing")
	}
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + qr["name"]}
	job, err := jobObj.Create(config, qr)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// Rename a job.
// First parameter job old name, Second parameter job new name.
func (j *Jenkins) RenameJob(job string, name string) *Job {
	jobObj := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + job}
	jobObj.Rename(name)
	return &jobObj
}

// Create a copy of a job.
// First parameter Name of the job to copy from, Second parameter new job name.
func (j *Jenkins) CopyJob(copyFrom string, newName string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + copyFrom}
	_, err := job.Poll()
	if err != nil {
		return nil, err
	}
	return job.Copy(newName)
}

// Delete a job.
func (j *Jenkins) DeleteJob(name string, parentIDs ...string) (bool, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, name), "/job/")}
	return job.Delete()
}

// Invoke a job.
// First parameter job name, second parameter is optional Build parameters.
func (j *Jenkins) BuildJob(name string, options ...interface{}) (int64, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + name}
	var params map[string]string
	if len(options) > 0 {
		params, _ = options[0].(map[string]string)
	}
	return job.InvokeSimple(params)
}

func (j *Jenkins) GetBuild(jobName string, number int64) (*Build, error) {
	job, err := j.GetJob(jobName)
	if err != nil {
		return nil, err
	}
	build, err := job.GetBuild(number)

	if err != nil {
		return nil, err
	}
	return build, nil
}

func (j *Jenkins) GetJob(id string, parentIDs ...string) (*Job, error) {
	job := Job{Jenkins: j, Raw: new(JobResponse), Base: "/job/" + strings.Join(append(parentIDs, id), "/job/")}
	status, err := job.Poll()
	if err != nil {
		return nil, err
	}
	if status == 200 {
		return &job, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

func (j *Jenkins) GetFolder(id string, parents ...string) (*Folder, error) {
	folder := Folder{Jenkins: j, Raw: new(FolderResponse), Base: "/job/" + strings.Join(append(parents, id), "/job/")}
	status, err := folder.Poll()
	if err != nil {
		return nil, fmt.Errorf("trouble polling folder: %v", err)
	}
	if status == 200 {
		return &folder, nil
	}
	return nil, errors.New(strconv.Itoa(status))
}

// Get all builds Numbers and URLS for a specific job.
// There are only build IDs here,
// To get all the other info of the build use jenkins.GetBuild(job,buildNumber)
// or job.GetBuild(buildNumber)

func (j *Jenkins) Poll() (int, error) {
	resp, err := j.Requester.GetJSON("/", j.Raw, nil)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

// Create a ssh credentials
// return credentials id
func (j *Jenkins) CreateSshCredential(id, username, passphrase, privateKey, description string) (*string, error) {
	requestStruct := NewCreateSshCredentialRequest(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	response, err := j.Requester.Post("/credentials/store/system/domain/_/createCredentials",
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateUsernamePasswordCredential(id, username, password, description string) (*string, error) {
	requestStruct := NewCreateUsernamePasswordRequest(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	response, err := j.Requester.Post("/credentials/store/system/domain/_/createCredentials",
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateSshCredentialRequest(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateUsernamePasswordRequest(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateSecretTextCredentialRequest(id, secret, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) CreateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error) {
	requestStruct := NewCreateKubeconfigCredentialRequest(id, content, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	responseString := ""
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/createCredentials", domain),
		nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &requestStruct.Credentials.Id, nil
}

func (j *Jenkins) UpdateSshCredentialInFolder(domain, id, username, passphrase, privateKey, description string, folders ...string) (*string, error) {
	requestStruct := NewSshCredential(id, username, passphrase, privateKey, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateUsernamePasswordCredentialInFolder(domain, id, username, password, description string, folders ...string) (*string, error) {
	requestStruct := NewUsernamePasswordCredential(id, username, password, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateSecretTextCredentialInFolder(domain, id, secret, description string, folders ...string) (*string, error) {
	requestStruct := NewSecretTextCredential(id, secret, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) UpdateKubeconfigCredentialInFolder(domain, id, content, description string, folders ...string) (*string, error) {
	requestStruct := NewKubeconfigCredential(id, content, description)
	param := map[string]string{"json": makeJson(requestStruct)}
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/updateSubmit", domain, id),
		nil, nil, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}






func (j *Jenkins) DeleteCredentialInFolder(domain, id string, folders ...string) (*string, error) {
	prePath := ""
	if domain == "" {
		domain = "_"
	}
	if len(folders) == 0 {
		return nil, fmt.Errorf("folder name shoud not be nil")
	}
	for _, folder := range folders {
		prePath = prePath + fmt.Sprintf("/job/%s", folder)
	}
	response, err := j.Requester.Post(prePath+
		fmt.Sprintf("/credentials/store/folder/domain/%s/credential/%s/doDelete", domain, id),
		nil, nil, nil)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return &id, nil
}

func (j *Jenkins) GetGlobalRole(roleName string) (*GlobalRole, error) {
	roleResponse := &GlobalRoleResponse{
		RoleName: roleName,
	}
	stringResponse := ""
	response, err := j.Requester.Get("/role-strategy/strategy/getRole",
		&stringResponse,
		map[string]string{
			"roleName": roleName,
			"type":     GLOBAL_ROLE,
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	if stringResponse == "{}" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(stringResponse), roleResponse)
	if err != nil {
		return nil, err
	}
	return &GlobalRole{
		Jenkins: j,
		Raw:     *roleResponse,
	}, nil
}

func (j *Jenkins) GetProjectRole(roleName string) (*ProjectRole, error) {
	roleResponse := &ProjectRoleResponse{
		RoleName: roleName,
	}
	stringResponse := ""
	response, err := j.Requester.Get("/role-strategy/strategy/getRole",
		&stringResponse,
		map[string]string{
			"roleName": roleName,
			"type":     PROJECT_ROLE,
		})
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	if stringResponse == "{}" {
		return nil, nil
	}
	err = json.Unmarshal([]byte(stringResponse), roleResponse)
	if err != nil {
		return nil, err
	}
	return &ProjectRole{
		Jenkins: j,
		Raw:     *roleResponse,
	}, nil
}

func (j *Jenkins) AddGlobalRole(roleName string, ids GlobalPermissionIds, overwrite bool) (*GlobalRole, error) {
	responseRole := &GlobalRole{
		Jenkins: j,
		Raw: GlobalRoleResponse{
			RoleName:      roleName,
			PermissionIds: ids,
		}}
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      roleName,
		"type":          GLOBAL_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(overwrite),
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseRole, nil
}

func (j *Jenkins) DeleteProjectRoles(roleName ...string) error {
	responseString := ""

	response, err := j.Requester.Post("/role-strategy/strategy/removeRoles", nil, &responseString, map[string]string{
		"type":      PROJECT_ROLE,
		"roleNames": strings.Join(roleName, ","),
	})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println(responseString)
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

func (j *Jenkins) AddProjectRole(roleName string, pattern string, ids ProjectPermissionIds, overwrite bool) (*ProjectRole, error) {
	responseRole := &ProjectRole{
		Jenkins: j,
		Raw: ProjectRoleResponse{
			RoleName:      roleName,
			PermissionIds: ids,
			Pattern:       pattern,
		}}
	var idArray []string
	values := reflect.ValueOf(ids)
	for i := 0; i < values.NumField(); i++ {
		field := values.Field(i)
		if field.Bool() {
			idArray = append(idArray, values.Type().Field(i).Tag.Get("json"))
		}
	}
	param := map[string]string{
		"roleName":      roleName,
		"type":          PROJECT_ROLE,
		"permissionIds": strings.Join(idArray, ","),
		"overwrite":     strconv.FormatBool(overwrite),
		"pattern":       pattern,
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/addRole", nil, &responseString, param)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New(strconv.Itoa(response.StatusCode))
	}
	return responseRole, nil
}

func (j *Jenkins) DeleteUserInProject(username string) error {
	param := map[string]string{
		"type": PROJECT_ROLE,
		"sid":  username,
	}
	responseString := ""
	response, err := j.Requester.Post("/role-strategy/strategy/deleteSid", nil, &responseString, param)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New(strconv.Itoa(response.StatusCode))
	}
	return nil
}

// Creates a new Jenkins Instance
// Optional parameters are: client, username, password
// After creating an instance call init method.
func CreateJenkins(client *http.Client, base string, maxConnection int, auth ...interface{}) *Jenkins {
	j := &Jenkins{}
	if strings.HasSuffix(base, "/") {
		base = base[:len(base)-1]
	}
	j.Server = base
	j.Requester = &Requester{Base: base, SslVerify: true, Client: client, connControl: make(chan struct{}, maxConnection)}
	if j.Requester.Client == nil {
		j.Requester.Client = http.DefaultClient
	}
	if len(auth) == 2 {
		j.Requester.BasicAuth = &BasicAuth{Username: auth[0].(string), Password: auth[1].(string)}
	}
	return j
}