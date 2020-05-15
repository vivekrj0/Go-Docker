package wrapper

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

var dockerHostRegexp = regexp.MustCompile("://([^:]+):")

func InferHost() (string, error) {
	envHost := os.Getenv("DOCKER_HOST")
	if len(envHost) == 0 {
		return "localhost", nil
	}

	matches := dockerHostRegexp.FindAllStringSubmatch(envHost, -1)
	if len(matches) != 1 || len(matches[0]) != 2 {
		return "", fmt.Errorf("Cannot parse DOCKER_HOST '%v'", envHost)
	}
	return matches[0][1], nil
}

func MustInferHost() string {
	dockerHost, err := InferHost()
	if err != nil {
		panic(err)
	}
	return dockerHost
}

func runCmd(name string, args ...string) (string, error) {
	var outBuf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf
	err := cmd.Run()
	out := outBuf.String()
	if err != nil {
		fmt.Print(out)
	}
	return out, err
}

func writeTmp(content string) (string, error) {
	f, err := ioutil.TempFile("", "docker-compose-")
	if err != nil {
		return "", fmt.Errorf("Error creating temp file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("Error writing temp file: %v", err)
	}

	return f.Name(), nil
}
