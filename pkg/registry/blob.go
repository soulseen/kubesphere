package registry

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	log "github.com/golang/glog"
)

// Digest returns the digest for an image.
func (r *Registry) Blob(image Image, token string) ([]byte, error) {

	url := r.GetBlobUrl(image)
	log.Info("registry.blobs.get url=" + url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", schema2.MediaTypeManifest)
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		log.Info("got status code: " + string(resp.StatusCode))
		return nil, fmt.Errorf("got status code: %d", resp.StatusCode)
	}
	respBody, _ := GetRespBody(resp)

	return respBody, nil
}

func (r *Registry) GetBlobUrl(image Image) string {
	url := r.url("/v2/%s/blobs/%s", image.Path, image.Digest)
	return url
}
