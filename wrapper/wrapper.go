package wrapper

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type Wrapper struct {
	fileName   string
	Containers map[string]*Container
}

var (
	logger           = log.New(os.Stdout, "Go-Docker: ", log.LstdFlags)
	replaceEnvRegexp = regexp.MustCompile("\\$\\{[^\\}]+\\}")
	composeUpRegexp  = regexp.MustCompile("(?m:docker start <- \\(u'(.*)'\\)$)")
)

const (
	composeProjectName = "godocker"
)

func Start(dockerWrapperYML string, forcePull, rmFirst bool) (*Wrapper, error) {
	logger.Println("initializing...")
	dockerWrapperYML = replaceEnv(dockerWrapperYML)

	fName, err := writeTmp(dockerWrapperYML)
	if err != nil {
		return nil, err
	}

	ids, err := godockerStart(fName, forcePull, rmFirst)
	if err != nil {
		return nil, err
	}

	containers := make(map[string]*Container)

	for _, id := range ids {
		container, err := Inspect(id)
		if err != nil {
			return nil, err
		}
		if !container.State.Running {
			return nil, fmt.Errorf("Container '%v' is not running", container.Name)
		}
		containers[container.Name[1:]] = container
	}

	return &Wrapper{fileName: fName, Containers: containers}, nil
}

func MustStart(dockerWrapperYML string, forcePull, killFirst bool) *Wrapper {
	godocker, err := Start(dockerWrapperYML, forcePull, killFirst)
	if err != nil {
		panic(err)
	}
	return godocker
}

func (c *Wrapper) Kill() error {
	return godockerKill(c.fileName)
}

func (c *Wrapper) MustKill() {
	if err := c.Kill(); err != nil {
		panic(err)
	}
}

func replaceEnv(dockerWrapperYML string) string {
	return replaceEnvRegexp.ReplaceAllStringFunc(dockerWrapperYML, replaceEnvFunc)
}

func replaceEnvFunc(s string) string {
	return os.Getenv(strings.TrimSpace(s[2 : len(s)-1]))
}

func godockerStart(fName string, forcePull, rmFirst bool) ([]string, error) {
	if forcePull {
		logger.Println("pulling images...")
		if _, err := godockerRun(fName, "pull"); err != nil {
			return nil, fmt.Errorf("Error pulling images: %v", err)
		}
	}

	if rmFirst {
		if err := godockerKill(fName); err != nil {
			return nil, err
		}
		if err := godockerRm(fName); err != nil {
			return nil, err
		}
	}

	logger.Println("starting containers...")
	out, err := godockerRun(fName, "--verbose", "up", "-d")
	if err != nil {
		return nil, fmt.Errorf("Error starting containers: %v", err)
	}
	logger.Println("Containers started")

	matches := composeUpRegexp.FindAllStringSubmatch(out, -1)
	ids := make([]string, 0, len(matches))
	for _, match := range matches {
		ids = append(ids, match[1])
	}

	return ids, nil
}

func godockerKill(fName string) error {
	logger.Println("killing stale containers...")
	_, err := godockerRun(fName, "kill")
	if err != nil {
		return fmt.Errorf("Error killing stale containers: %v", err)
	}
	return err
}

func godockerRm(fName string) error {
	logger.Println("removing stale containers...")
	_, err := godockerRun(fName, "rm", "--force")
	if err != nil {
		return fmt.Errorf("Error removing stale containers: %v", err)
	}
	return err
}

func godockerRun(fName string, otherArgs ...string) (string, error) {
	args := []string{"-f", fName, "-p", composeProjectName}
	args = append(args, otherArgs...)
	return runCmd("docker-compose", args...)
}
