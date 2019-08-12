package registry

import (
	"testing"
)

func TestDigestFromDockerHub(t *testing.T) {

	testImage := Image{Domain: "docker.io", Path: "library/alpine", Tag: "latest"}
	r, err := CreateRegistryClient("", "", "", "docker.io")
	if err != nil {
		t.Fatalf("Could not get client: %s", err)
	}

	digestUrl := r.GetDigestUrl(testImage)

	// Get token.
	token, err := r.Token(digestUrl)
	if err != nil || token == "" {
		t.Fatalf("Could not get token: %s", err)
	}

	d, err := r.Digest(testImage, token)
	if err != nil {
		t.Fatalf("Could not get digest: %s", err)
	}

	if d == "" {
		t.Error("Empty digest received")
	}
}
