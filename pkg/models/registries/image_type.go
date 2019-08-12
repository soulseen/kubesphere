package registries

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type AuthInfo struct {
	Username   string `json:"username" description:"username"`
	Password   string `json:"password" description:"password"`
	ServerHost string `json:"serverhost" description:"registry server host"`
}

type ImageNameAndSecret struct {
	ImageName string        `json:"imageName,omitempty" description:"Image full name."`
	Secret    corev1.Secret `json:"secret,omitempty" description:"Secret holds secret data of a certain type."`
}

// ImageBlobInfo describes the info of an image.
type ImageBlobInfo struct {
	// Status is the status of the image search, such as "failed","succeeded".
	Status string `json:"status" description:"Status is the status of the image search, such as \"failed\",\"succeeded\"."`
	// Image is the metadata of that image
	Image *ImageBlob `json:"imageBlob,omitempty" description:"Retrieve the blob from the registry identified. Reference: https://docs.docker.com/registry/spec/api/#blob"`
}

type ImageBlob struct {
	Architecture    string          `json:"architecture,omitempty" description:"The architecture field specifies the CPU architecture, for example amd64 or ppc64le."`
	Config          Config          `json:"config,omitempty" description:"The config field references a configuration object for a container."`
	Container       string          `json:"container,omitempty" description:"Container id."`
	ContainerConfig ContainerConfig `json:"container_config,omitempty" description:"The config data of container."`
	Created         time.Time       `json:"created,omitempty" description:"Create time."`
	DockerVersion   string          `json:"docker_version,omitempty" description:"docker version."`
	History         []History       `json:"history,omitempty" description:"The data of history update."`
	Os              string          `json:"os,omitempty" description:"Operating system."`
	Rootfs          Rootfs          `json:"rootfs omitempty" description:"Root filesystem."`
}

type Labels struct {
	Maintainer string `json:"maintainer" description:""`
}
type Config struct {
	HostName     string                 `json:"Hostname,omitempty" description:"A string value containing the hostname to use for the container."`
	DomainName   string                 `json:"Domainname,omitempty" description:"A string value containing the domain name to use for the container."`
	User         string                 `json:"User,omitempty" description:"A string value specifying the user inside the container."`
	AttachStdin  bool                   `json:"AttachStdin,omitempty" description:"Boolean value, attaches to stdin."`
	AttachStdout bool                   `json:"AttachStdout,omitempty" description:"Boolean value, attaches to stdout."`
	AttachStderr bool                   `json:"AttachStderr,omitempty" description:"Boolean value, attaches to stderr."`
	ExposedPorts map[string]interface{} `json:"ExposedPorts,omitempty" description:"An object mapping ports to an empty object in the form of: \"ExposedPorts\": { \"<port>/<tcp|udp>: {}\" }"`
	Tty          bool                   `json:"Tty,omitempty" description:"Boolean value, Attach standard streams to a tty, including stdin if it is not closed."`
	OpenStdin    bool                   `json:"OpenStdin,omitempty" description:"Boolean value, opens stdin"`
	StdinOnce    bool                   `json:"StdinOnce,omitempty" description:"Boolean value, close stdin after the 1 attached client disconnects."`
	Env          []string               `json:"Env,omitempty" description:"A list of environment variables in the form of [\"VAR=value\", ...]"`
	Cmd          []string               `json:"Cmd,omitempty" description:"Command to run specified as a string or an array of strings."`
	ArgsEscaped  bool                   `json:"ArgsEscaped,omitempty" description:"Command is already escaped (Windows only)"`
	Image        string                 `json:"Image,omitempty" description:"A string specifying the image name to use for the container."`
	Volumes      interface{}            `json:"Volumes,omitempty" description:"An object mapping mount point paths (strings) inside the container to empty objects."`
	WorkingDir   string                 `json:"WorkingDir,omitempty" description:"A string specifying the working directory for commands to run in."`
	Entrypoint   interface{}            `json:"Entrypoint,omitempty" description:"The entry point set for the container as a string or an array of strings."`
	OnBuild      interface{}            `json:"OnBuild,omitempty" description:"ONBUILD metadata that were defined in the image's Dockerfile."`
	Labels       Labels                 `json:"Labels,omitempty" description:"The map of labels to a container."`
	StopSignal   string                 `json:"StopSignal,omitempty" description:"Signal to stop a container as a string or unsigned integer."`
}
type ContainerConfig struct {
	HostName     string                 `json:"Hostname,omitempty" description:"A string value containing the hostname to use for the container."`
	DomainName   string                 `json:"Domainname,omitempty" description:"A string value containing the domain name to use for the container."`
	User         string                 `json:"User,omitempty" description:"A string value specifying the user inside the container."`
	AttachStdin  bool                   `json:"AttachStdin,omitempty" description:"Boolean value, attaches to stdin."`
	AttachStdout bool                   `json:"AttachStdout,omitempty" description:"Boolean value, attaches to stdout."`
	AttachStderr bool                   `json:"AttachStderr,omitempty" description:"Boolean value, attaches to stderr."`
	ExposedPorts map[string]interface{} `json:"ExposedPorts,omitempty" description:"An object mapping ports to an empty object in the form of: \"ExposedPorts\": { \"<port>/<tcp|udp>: {}\" }"`
	Tty          bool                   `json:"Tty,omitempty" description:"Boolean value, Attach standard streams to a tty, including stdin if it is not closed."`
	OpenStdin    bool                   `json:"OpenStdin,omitempty" description:"Boolean value, opens stdin"`
	StdinOnce    bool                   `json:"StdinOnce,omitempty" description:"Boolean value, close stdin after the 1 attached client disconnects."`
	Env          []string               `json:"Env,omitempty" description:"A list of environment variables in the form of [\"VAR=value\", ...]"`
	Cmd          []string               `json:"Cmd,omitempty" description:"Command to run specified as a string or an array of strings."`
	ArgsEscaped  bool                   `json:"ArgsEscaped,omitempty" description:"Command is already escaped (Windows only)"`
	Image        string                 `json:"Image,omitempty" description:"A string specifying the image name to use for the container."`
	Volumes      interface{}            `json:"Volumes,omitempty" description:"An object mapping mount point paths (strings) inside the container to empty objects."`
	WorkingDir   string                 `json:"WorkingDir,omitempty" description:"A string specifying the working directory for commands to run in."`
	EntryPoint   interface{}            `json:"Entrypoint,omitempty" description:"The entry point set for the container as a string or an array of strings."`
	OnBuild      interface{}            `json:"OnBuild,omitempty" description:"ONBUILD metadata that were defined in the image's Dockerfile."`
	Labels       Labels                 `json:"Labels,omitempty" description:"The map of labels to a container."`
	StopSignal   string                 `json:"StopSignal,omitempty" description:"Signal to stop a container as a string or unsigned integer."`
}
type History struct {
	Created    time.Time `json:"created,omitempty" description:"Created time."`
	CreatedBy  string    `json:"created_by,omitempty" description:"Created command."`
	EmptyLayer bool      `json:"empty_layer,omitempty" description:"Layer empty or not."`
}
type Rootfs struct {
	Type    string   `json:"type,omitempty" description:"Root filesystem type, always \"layers\" "`
	DiffIds []string `json:"diff_ids,omitempty" description:"Contain ids of layer list"`
}
