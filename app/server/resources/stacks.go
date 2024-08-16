package resources

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"github.com/fatih/structs"
)

// Represent a Docker stack
type Stack struct {
	Name        string
	Status      string
	ConfigFiles string
}

// Represent an array of Docker stacks
type Stacks []Stack

// Retrieve all inspector tabs for Docker stacks
func StacksInspectorTabs() []string {
	return []string{"Logs", "Services", "Config"}
}

// Retrieve all the single actions associated with Docker stacks
func StackSingleActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:            "up the stack",
			Command:          "stack.up",
			Key:              "u",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "pause/unpause the stack",
			Command:          "stack.pause",
			Key:              "p",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "down the stack",
			Command:          "stack.down",
			Key:              "d",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "restart the stack",
			Command:          "stack.restart",
			Key:              "r",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "update the stack (down, pull, up)",
			Command:          "stack.update",
			Key:              "U",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "edit the stack configuration",
			Command:          "stack.update",
			Key:              "e",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "create a new stack",
			Command:          "createStack",
			Key:              "C",
			RequiresResource: false,
			RunLocally:       true,
		},
	)

	return actions
}

// Retrieve all the bulk actions associated with Docker stacks
func StacksBulkActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "update all stacks",
			Prompt:  "Are you sure you want to update all stacks?",
			Command: "stacks.update",
		},
	)
	return actions
}

// Retrieve all Docker stacks
func StacksList(client *client.Client) Stacks {
	if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
		return []Stack{}
	}

	output, err := exec.Command("docker", "compose", "ls", "--format", "json").Output()

	if err != nil {
		return []Stack{}
	}

	var stacks []Stack
	err = json.Unmarshal(output, &stacks)

	if err != nil {
		return []Stack{}
	}

	return stacks
}

// Count the number of Docker stacks
func StacksCount(client *client.Client) int {
	var list = StacksList(client)
	return len(list)
}

// Turn the list of Docker stacks into a list of rows representing them
func (stacks Stacks) ToRows(columns []string) ui.Rows {
	var rows = make(ui.Rows, 0)

	sort.Slice(stacks, func(i, j int) bool {
		return stacks[i].Name < stacks[j].Name
	})

	for i := 0; i < len(stacks); i++ {
		stack := stacks[i]

		row := structs.Map(stack)
		var flat = make([]map[string]string, 0)

		for j := 0; j < len(columns); j++ {
			_entry := make(map[string]string)
			_entry["field"] = columns[j]

			switch columns[j] {
			case "Name":
				_entry["value"] = stack.Name
			case "Status":
				_entry["value"] = strings.Split(stack.Status, "(")[0]
			case "ConfigFiles":
				_entry["value"] = stack.ConfigFiles
			}

			flat = append(flat, _entry)
		}
		row["_representation"] = flat
		rows = append(rows, row)
	}

	return rows
}

// Single - Start the stack (docker compose up -d)
func (s Stack) Up(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "up", "-d").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Single - Pause the stack (docker compose pause)
func (s Stack) Pause(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "pause").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Single - Unpause the stack (docker compose unpause)
func (s Stack) Unpause(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "unpause").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Single - Down the stack (docker compose down)
func (s Stack) Down(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "down").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Single - Update the stack (docker compose down, docker compose pull, docker compose up)
func (s Stack) Update(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "down").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	output, err = exec.Command("docker", "compose", "-f", s.ConfigFiles, "pull").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	output, err = exec.Command("docker", "compose", "-f", s.ConfigFiles, "up", "-d").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Single - Restart the stack (docker compose restart)
func (s Stack) Restart(client *client.Client) error {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "restart").CombinedOutput()

	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

// Inspector - Retrieve the list of services (containers) inside a Docker stack
func (s Stack) GetServices(client *client.Client) (ui.InspectorContent, error) {
	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "ps", "-aq").Output()

	if err != nil {
		return nil, err
	}

	ids := strings.Split(string(output), "\n")
	filterArgs := filters.NewArgs()
	for _, id := range ids {
		filterArgs.Add("id", id)
	}

	containers := ContainersList(client, filterArgs)

	allConfig := ui.InspectorContent{
		ui.InspectorContentPart{Type: "rows", Content: containers.ToRows(strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ","))},
	}

	return allConfig, nil
}

// Inspector - Retrieve the full configuration associated with a Docker stack
func (s Stack) GetConfig(client *client.Client) (ui.InspectorContent, error) {
	firstPartRows := make(ui.Rows, 0)
	firstPartRows = append(firstPartRows, ui.Row{"_representation": []string{"Location:", s.ConfigFiles}})
	firstPart := ui.InspectorContentPart{Type: "rows", Content: firstPartRows}

	separator := ui.InspectorContentPart{Type: "lines", Content: []string{}}

	config, err := os.ReadFile(s.ConfigFiles)

	if err != nil {
		return nil, err
	}

	code := strings.Split(string(config), "\n")

	allConfig := ui.InspectorContent{
		firstPart,
		separator,
		ui.InspectorContentPart{Type: "code", Content: code},
	}

	return allConfig, nil
}

// Inspector - Retrieve the full configuration associated with a Docker stack - The raw file lines only
func (s Stack) GetRawConfig(client *client.Client) (string, error) {
	config, err := os.ReadFile(s.ConfigFiles)

	if err != nil {
		return "", err
	}

	return string(config), nil
}

// Inspector - Retrieve the logs written by the Docker stack
func (s Stack) GetLogs(client *client.Client, writer io.Writer, showTimestamps bool) (*io.ReadCloser, error) {
	opts := make([]string, 0)

	opts = append(opts, "compose")
	opts = append(opts, "-f")
	opts = append(opts, s.ConfigFiles)
	opts = append(opts, "logs")

	opts = append(opts, "--follow")
	opts = append(opts, "--no-color")

	opts = append(opts, "--since")
	opts = append(opts, _os.GetEnv("CONTAINER_LOGS_SINCE"))

	opts = append(opts, "--tail")
	opts = append(opts, _os.GetEnv("CONTAINER_LOGS_TAIL"))

	if showTimestamps {
		opts = append(opts, "--timestamps")
	}

	process := exec.Command("docker", opts...)
	reader, err := process.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = process.Start()
	if err != nil {
		return nil, err
	}

	go io.Copy(writer, reader)

	return &reader, nil
}

// Create a new Docker stack from a docker-compose.yml content
func StackCreate(c *client.Client, m process.LongTaskMonitor, args map[string]interface{}) {
	content := args["Content"].(string)
	filename := fmt.Sprintf("docker-compose.%s.yml", uuid.NewString())
	filepath := path.Join(_os.GetEnv("STACKS_DIRECTORY"), filename)

	err := os.WriteFile(filepath, []byte(content), 0644)

	if err != nil {
		m.Errors <- err
		return
	}

	output, err := exec.Command("docker", "compose", "-f", filepath, "config").CombinedOutput()

	if err != nil {
		m.Errors <- errors.New(string(output))
		return
	}

	process := exec.Command("docker", "compose", "-f", filepath, "up", "-d")
	reader, err := process.StdoutPipe()

	if err != nil {
		m.Errors <- err
		return
	}

	err = process.Start()
	if err != nil {
		m.Errors <- err
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			m.Results <- scanner.Text()
		}
		wg.Done()
	}()

	wg.Wait()
	m.Done <- true
}

// Edit an existing Docker stack by overwriting a docker-compose.yml (down, overwrite, up)
func (s Stack) Edit(c *client.Client, m process.LongTaskMonitor, args map[string]interface{}) {
	content := args["Content"].(string)
	err := s.Down(c)

	if err != nil {
		m.Errors <- err
		return
	}

	originalContent, err := os.ReadFile(s.ConfigFiles)

	if err != nil {
		m.Errors <- err
		return
	}

	err = os.WriteFile(s.ConfigFiles, []byte(content), 0644)

	if err != nil {
		m.Errors <- err
		return
	}

	output, err := exec.Command("docker", "compose", "-f", s.ConfigFiles, "config").CombinedOutput()

	if err != nil {
		m.Errors <- errors.New(string(output))
		os.WriteFile(s.ConfigFiles, originalContent, 0644)
		s.Up(c)
		return
	}

	process := exec.Command("docker", "compose", "-f", s.ConfigFiles, "up", "-d")
	reader, err := process.StdoutPipe()

	if err != nil {
		m.Errors <- err
		return
	}

	err = process.Start()
	if err != nil {
		m.Errors <- err
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			m.Results <- scanner.Text()
		}
		wg.Done()
	}()

	wg.Wait()
	m.Done <- true
}
