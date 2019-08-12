package registry

import (
	"github.com/docker/docker/api/types"
	log "github.com/golang/glog"
)

const (
	// DefaultDockerRegistry is the default docker registry address.
	DefaultDockerRegistry = "https://registry-1.docker.io"
)

func CreateRegistryClient(authURL, username, password, domain string) (*Registry, error) {
	authDomain := authURL
	if authDomain == "" {
		authDomain = domain
	}

	auth, err := GetAuthConfig(username, password, authDomain)
	if err != nil {
		return nil, err
	}

	// Create the registry client.
	log.Infof("domain: %s", domain)
	log.Infof("server address: %s", auth.ServerAddress)

	return New(auth, Opt{
		Domain: domain,
	})
}

// GetAuthConfig returns the docker registry AuthConfig.
func GetAuthConfig(username, password, registry string) (types.AuthConfig, error) {
	registry = setDefaultRegistry(registry)
	if username != "" && password != "" && registry != "" {
		return types.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: registry,
		}, nil
	}

	log.Info("Using registry ", registry, " with no authentication")

	return types.AuthConfig{
		ServerAddress: registry,
	}, nil

}

func setDefaultRegistry(serverAddress string) string {
	if serverAddress == "docker.io" || serverAddress == "" {
		serverAddress = DefaultDockerRegistry
	}

	return serverAddress
}
