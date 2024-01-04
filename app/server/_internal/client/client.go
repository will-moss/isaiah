package client

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/client"

	_os "will-moss/isaiah/server/_internal/os"
)

// Alias for client.NewClientWithOpts, without returning any error
func NewClientWithOpts(ops client.Opt) *client.Client {
	_client, _ := client.NewClientWithOpts(ops)
	return _client
}

// Try to find the current Docker host on the system, using :
// 1. Env variable : CUSTOMER_DOCKER_HOST
// 2. Env variable : DOCKER_HOST
// 3. Env variable : DOCKER_CONTEXT
// 4. Output of command : docker context show + docker context inspect
// 5. OS-based default location
func DiscoverDockerHost() (string, error) {
	// 1. Custom Docker host provided
	if _os.GetEnv("CUSTOM_DOCKER_HOST") != "" {
		return _os.GetEnv("CUSTOM_DOCKER_HOST"), nil
	}

	// 2. Default Docker host already set
	if _os.GetEnv("DOCKER_HOST") != "" {
		return _os.GetEnv("DOCKER_HOST"), nil
	}

	if _os.GetEnv("DOCKER_RUNNING") != "TRUE" {
		// 3. Default Docker context already set
		if _os.GetEnv("DOCKER_CONTEXT") != "" {
			cmd := exec.Command("docker", "context", "inspect", _os.GetEnv("DOCKER_CONTEXT"))
			output, err := cmd.Output()

			if err != nil {
				return "", fmt.Errorf("An error occurred while trying to inspect the Docker context provided : %s", err)
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "Host") {
					parts := strings.Split(line, "\"Host\": ")
					replacer := strings.NewReplacer("\"", "", ",", "")
					host := replacer.Replace(parts[1])

					return host, nil
				}
			}
		}

		// 4. Attempt to retrieve the current Docker context if all the other cases proved unsuccesful
		{
			cmd := exec.Command("docker", "context", "show")
			output, err := cmd.Output()

			if err != nil {
				return "", fmt.Errorf("An error occurred while trying to retrieve the default Docker context : %s", err)
			}

			currentContext := strings.TrimSpace(string(output))
			if currentContext != "" {
				cmd := exec.Command("docker", "context", "inspect", currentContext)

				output, err := cmd.Output()
				if err != nil {
					return "", fmt.Errorf("An error occurred while trying to inspect the default Docker context : %s", err)
				}

				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					if strings.Contains(line, "Host") {
						parts := strings.Split(line, "\"Host\": ")
						replacer := strings.NewReplacer("\"", "", ",", "")
						host := replacer.Replace(parts[1])
						return host, nil
					}
				}
			}
		}
	}

	// 5. Every previous attempt failed, try to use the default location
	// 5.1. Unix-like systems
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		return "unix:///var/run/docker.sock", nil
	}
	// 5.2. Windows system
	if _, err := os.Stat("\\\\.\\pipe\\docker_engine"); err == nil {
		return "\\\\.\\pipe\\docker_engine", nil
	}

	var finalError error
	if _os.GetEnv("DOCKER_RUNNING") != "TRUE" {
		finalError = fmt.Errorf("Automatic Docker host discovery failed on your system. Please try setting DOCKER_HOST manually")
	} else {
		finalError = fmt.Errorf("Automatic Docker host discovery failed on your system. Please make sure your Docker socket is mounted on your container")
	}
	return "", finalError
}
