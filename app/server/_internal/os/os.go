package os

import (
	"os"
	"os/exec"
	"strings"
	"will-moss/isaiah/server/_internal/tty"
)

// Alias for os.GetEnv, with support for fallback value, and boolean normalization
func GetEnv(key string, fallback ...string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		if len(fallback) > 0 {
			value = fallback[0]
		} else {
			value = ""
		}
	} else {
		// Quotes removal
		value = strings.Trim(value, "\"")

		// Boolean normalization
		mapping := map[string]string{
			"0":     "FALSE",
			"off":   "FALSE",
			"false": "FALSE",
			"1":     "TRUE",
			"on":    "TRUE",
			"true":  "TRUE",
			"rue":   "TRUE",
		}
		normalized, isBool := mapping[strings.ToLower(value)]
		if isBool {
			value = normalized
		}
	}

	return value
}

// Retrieve all the environment variables as a map
func GetFullEnv() map[string]string {
	var structured = make(map[string]string)

	raw := os.Environ()
	for i := 0; i < len(raw); i++ {
		pair := strings.Split(raw[i], "=")
		key := pair[0]
		value := GetEnv(key)

		structured[key] = value
	}
	return structured
}

// Open a shell on the system, and update the provided channels with 
// status / errors as they happen
func OpenShell(tty *tty.TTY, channelErrors chan error, channelUpdates chan string) {
	cmd := GetEnv("TTY_SERVER_COMMAND")
	cmdParts := strings.Split(cmd, " ")

	process := exec.Command(cmdParts[0], cmdParts[1:]...)
	process.Stdin = tty.Stdin
	process.Stderr = tty.Stdout
	process.Stdout = tty.Stdout
	err := process.Start()

	if err != nil {
		channelErrors <- err
	} else {
		channelUpdates <- "started"
		process.Wait()
		channelUpdates <- "exited"
	}
}
