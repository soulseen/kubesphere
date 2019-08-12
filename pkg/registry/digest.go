package registry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"

	log "github.com/golang/glog"
)

// Digest returns the digest for an image.
func (r *Registry) Digest(image Image, token string) (digest.Digest, error) {
	if len(image.Digest) > 1 {
		// return early if we already have an image digest.
		return image.Digest, nil
	}

	url := r.GetDigestUrl(image)
	log.Info("registry.manifests.get url=" + url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return "", fmt.Errorf("got status code: %d", resp.StatusCode)
	}

	imageDigest := &ImageDigest{}

	respBody, _ := GetRespBody(resp)
	json.Unmarshal(respBody, imageDigest)

	return imageDigest.Config.Digest, nil
}

func (r *Registry) GetDigestUrl(image Image) string {
	url := r.url("/v2/%s/manifests/%s", image.Path, image.Tag)
	return url
}

type ImageDigest struct {
	SchemaVersion int      `json:"schemaVersion"`
	MediaType     string   `json:"mediaType"`
	Config        Config   `json:"config"`
	Layers        []Layers `json:"layers"`
}
type Config struct {
	MediaType string        `json:"mediaType"`
	Size      int           `json:"size"`
	Digest    digest.Digest `json:"digest"`
}
type Layers struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}
