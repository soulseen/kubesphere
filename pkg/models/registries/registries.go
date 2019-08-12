/*

 Copyright 2019 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

*/
package registries

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/registry"

	log "github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
)

const (
	loginSuccess  = "Login Succeeded"
	statusFailed  = "failed"
	statusSuccess = "succeeded"
)

type DockerConfigJson struct {
	Auths DockerConfigMap `json:"auths"`
}

// DockerConfig represents the config file used by the docker CLI.
// This config that represents the credentials that should be used
// when pulling images from specific image repositories.
type DockerConfigMap map[string]DockerConfigEntry

type DockerConfigEntry struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ServerAddress string `json:"serverAddress,omitempty"`
}

func RegistryVerify(authInfo AuthInfo) error {
	auth := base64.StdEncoding.EncodeToString([]byte(authInfo.Username + ":" + authInfo.Password))
	ctx := context.Background()
	cli, err := client.NewEnvClient()

	if err != nil {
		glog.Error(err)
	}

	config := types.AuthConfig{
		Username:      authInfo.Username,
		Password:      authInfo.Password,
		Auth:          auth,
		ServerAddress: authInfo.ServerHost,
	}

	resp, err := cli.RegistryLogin(ctx, config)
	cli.Close()

	if err != nil {
		return err
	}

	if resp.Status == loginSuccess {
		return nil
	} else {
		return fmt.Errorf(resp.Status)
	}
}

func RegistryImageBlob(imageNameAndSecret ImageNameAndSecret) (*ImageBlobInfo, error) {

	entry, err := getDockerEntryFromDockerSecret(&imageNameAndSecret.Secret)

	image, err := registry.ParseImage(imageNameAndSecret.ImageName)

	// Create the registry client.
	r, err := registry.CreateRegistryClient("", entry.Username, entry.Password, image.Domain)
	if err != nil {
		log.Error(err)
		return &ImageBlobInfo{Status: statusFailed}, nil
	}

	digestUrl := r.GetDigestUrl(image)

	// Get token.
	token, err := r.Token(digestUrl)

	// Get digest.
	digest, err := r.Digest(image, token)
	if err != nil {
		log.Error(err)
		return &ImageBlobInfo{Status: statusFailed}, nil
	}
	image.Digest = digest

	// Get blob.
	blob, err := r.Blob(image, token)
	if err != nil {
		log.Error(err)
		return &ImageBlobInfo{Status: statusFailed}, nil
	}

	imageBlob := &ImageBlob{}
	err = json.Unmarshal(blob, imageBlob)
	if err != nil {
		log.Error(err)
		return &ImageBlobInfo{Status: statusFailed}, nil
	}
	return &ImageBlobInfo{Status: statusSuccess, Image: imageBlob}, nil

}

func getDockerEntryFromDockerSecret(instance *corev1.Secret) (dockerConfigEntry *DockerConfigEntry, err error) {
	if instance.Type == "" {
		entry := &DockerConfigEntry{
			Username: "",
			Password: "",
		}
		return entry, nil
	}

	if instance.Type != corev1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("secret %s in ns %s type should be %s",
			instance.Namespace, instance.Name, corev1.SecretTypeDockerConfigJson)
	}
	dockerConfigBytes, ok := instance.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("could not get data %s", corev1.DockerConfigJsonKey)
	}
	dockerConfig := &DockerConfigJson{}
	err = json.Unmarshal(dockerConfigBytes, dockerConfig)
	if err != nil {
		return nil, err
	}
	if len(dockerConfig.Auths) == 0 {
		return nil, fmt.Errorf("docker config auth len should not be 0")
	}
	for registryAddress, dockerConfigEntry := range dockerConfig.Auths {
		dockerConfigEntry.ServerAddress = registryAddress
		return &dockerConfigEntry, nil
	}
	return nil, nil
}
