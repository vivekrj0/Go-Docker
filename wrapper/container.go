package wrapper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Container struct {
	ID              string           `json:"Id"`
	Name            string           `json:"Name,omitempty"`
	Created         time.Time        `json:"Created,omitempty"`
	Config          *Config          `json:"Config,omitempty"`
	State           State            `json:"State,omitempty"`
	Image           string           `json:"Image,omitempty"`
	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty"`
}

type Config struct {
	Hostname     string              `json:"Hostname,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	Cmd          []string            `json:"Cmd"`
	Image        string              `json:"Image,omitempty"`
	Labels       map[string]string   `json:"Labels,omitempty"`
}

type State struct {
	Running    bool      `json:"Running,omitempty"`
	Paused     bool      `json:"Paused,omitempty"`
	Restarting bool      `json:"Restarting,omitempty"`
	OOMKilled  bool      `json:"OOMKilled,omitempty"`
	Pid        int       `json:"Pid,omitempty"`
	ExitCode   int       `json:"ExitCode,omitempty"`
	Error      string    `json:"Error,omitempty"`
	StartedAt  time.Time `json:"StartedAt,omitempty"`
	FinishedAt time.Time `json:"FinishedAt,omitempty"`
}

type NetworkSettings struct {
	Ports map[string][]PortBinding `json:"Ports,omitempty"`
}

type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty"`
}

func Inspect(id string) (*Container, error) {
	out, err := runCmd("docker", "inspect", id)
	if err != nil {
		return nil, fmt.Errorf("Error inspecting container: %v", err)
	}

	var inspect []*Container
	if err := json.Unmarshal([]byte(out), &inspect); err != nil {
		return nil, fmt.Errorf("Error parsing inspect output: %v", err)
	}
	if len(inspect) != 1 {
		return nil, fmt.Errorf("Inspect returned %v results, 1 expected", len(inspect))
	}

	return inspect[0], nil
}

func MustInspect(id string) *Container {
	container, err := Inspect(id)
	if err != nil {
		panic(err)
	}
	return container
}

func (c *Container) GetPublicPort(exposedPort uint32, proto string) (uint32, error) {
	if c.NetworkSettings == nil {
		return 0, fmt.Errorf("No network settings for container '%v'", c.Name)
	}

	portSpec := fmt.Sprintf("%v/%v", exposedPort, strings.ToLower(proto))
	mapping, ok := c.NetworkSettings.Ports[portSpec]
	if !ok || len(mapping) == 0 {
		return 0, fmt.Errorf("No public port for %v", portSpec)
	}

	port, err := strconv.ParseUint(mapping[0].HostPort, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("Error parsing port '%v'", mapping[0].HostPort)
	}

	return uint32(port), nil
}

func (c *Container) MustGetPublicPort(exposedPort uint32, proto string) uint32 {
	port, err := c.GetPublicPort(exposedPort, proto)
	if err != nil {
		panic(err)
	}
	return port
}
