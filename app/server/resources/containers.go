package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/fatih/structs"
)

// Represent a Docker container
type Container struct {
	ID       string
	State    string
	ExitCode int
	Name     string
	Image    string
	Ports    []types.Port
	Created  int64
}

// Represent an  array of Docker containers
type Containers []Container

// Status translations using one/two-letter words
var shortStateTranslations = map[string]string{
	"paused":     "P",
	"exited":     "X",
	"created":    "C",
	"removing":   "RM",
	"restarting": "RS",
	"running":    "R",
	"dead":       "D",
}

// Status translations using symbol icons
var iconStateTranslations = map[string]rune{
	"paused":     '◫',
	"exited":     '⨯',
	"created":    '+',
	"removing":   '−',
	"restarting": '⟳',
	"running":    '▶',
	"dead":       '!',
}

// Retrieve all inspector tabs for Docker containers
func ContainersInspectorTabs() []string {
	return []string{"Logs", "Stats", "Env", "Config", "Top"}
}

// Retrieve all the single actions associated with Docker containers
func ContainerSingleActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "d",
			Label:            "remove container",
			Command:          "container.menu.remove",
			RequiresResource: true,
		},
		ui.MenuAction{
			Key:              "p",
			Label:            "pause/unpause container",
			Command:          "container.pause",
			RequiresResource: true,
		},
		ui.MenuAction{
			Key:              "s",
			Label:            "stop container",
			Command:          "container.stop",
			Prompt:           "Are you sure you want to stop this container?",
			RequiresResource: true,
		},
		ui.MenuAction{
			Key:              "r",
			Label:            "restart container",
			Command:          "container.restart",
			Prompt:           "Are you sure you want to restart this container?",
			RequiresResource: true,
		},
		ui.MenuAction{
			Key:              "m",
			Label:            "rename container",
			Command:          "rename",
			RequiresResource: false,
			RunLocally:       true,
		},
		ui.MenuAction{
			Key:              "E",
			Label:            "exec shell inside container",
			Command:          "container.shell",
			RequiresResource: true,
		},
		ui.MenuAction{
			Key:              "w",
			Label:            "open in browser",
			Command:          "container.browser",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the remove actions associated with Docker containers
func ContainerRemoveActions(c Container) []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove",
			Label:            fmt.Sprintf("<em>docker rm %s</em> ?", c.Name),
			Command:          "container.remove.default",
			RequiresResource: true,
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove with volumes",
			Label:            fmt.Sprintf("<em>docker rm --volumes %s</em> ?", c.Name),
			Command:          "container.remove.default.volumes",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the bulk actions associated with Docker containers
func ContainersBulkActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "stop all containers",
			Prompt:  "Are you sure you want to stop all containers?",
			Command: "containers.stop",
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "remove all containers (forced)",
			Prompt:  "Are you sure you want to remove all containers?",
			Command: "containers.remove",
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "prune unused containers",
			Prompt:  "Are you sure you want to prune all unused containers?",
			Command: "containers.prune",
		},
	)
	return actions
}

// Retrieve all Docker containers
func ContainersList(client *client.Client) Containers {
	reader, err := client.ContainerList(context.Background(), types.ContainerListOptions{All: true})

	if err != nil {
		return []Container{}
	}

	var containers []Container
	for i := 0; i < len(reader); i++ {
		var information = reader[i]

		var container Container
		container.ID = information.ID
		container.Name = information.Names[0][1:]
		container.State = information.State
		container.Image = information.Image
		container.Ports = information.Ports
		container.Created = information.Created

		inspection, err := client.ContainerInspect(context.Background(), information.ID)
		if err == nil {
			container.ExitCode = inspection.State.ExitCode
		}

		containers = append(containers, container)
	}

	return containers
}

// Stop all Docker containers
func ContainersStop(client *client.Client, monitor process.LongTaskMonitor, args map[string]interface{}) {
	containers := ContainersList(client)

	wg := sync.WaitGroup{}
	wg.Add(len(containers))

	for i := 0; i < len(containers); i++ {
		go func(_container Container) {
			defer wg.Done()

			err := client.ContainerStop(context.Background(), _container.ID, container.StopOptions{})
			if err != nil {
				monitor.Errors <- err
				return
			}
			monitor.Results <- _container.ID
		}(containers[i])
	}

	wg.Wait()
	monitor.Done <- true
}

// Force remove Docker containers
func ContainersRemove(client *client.Client) error {
	containers := ContainersList(client)

	for i := 0; i < len(containers); i++ {
		_container := containers[i]

		err := client.ContainerRemove(context.Background(), _container.ID, types.ContainerRemoveOptions{Force: true})

		if err != nil {
			return err
		}
	}

	return nil
}

// Prune unused Docker containers
func ContainersPrune(client *client.Client) error {
	_, err := client.ContainersPrune(context.Background(), filters.Args{})
	return err
}

// Turn the list of Docker containers into a list of rows representing them
func (containers Containers) ToRows(columns []string) ui.Rows {
	var rows = make(ui.Rows, 0)

	sort.Slice(containers, func(i, j int) bool {
		if containers[i].State == "running" && containers[j].State != "running" {
			return true
		}
		if containers[j].State == "running" && containers[i].State != "running" {
			return false
		}

		return containers[i].Name < containers[j].Name
	})

	for i := 0; i < len(containers); i++ {
		container := containers[i]

		row := structs.Map(container)
		var flat = make([]map[string]string, 0)

		for j := 0; j < len(columns); j++ {
			_entry := make(map[string]string)
			_entry["field"] = columns[j]

			switch columns[j] {
			case "ID":
				_entry["value"] = container.ID
			case "Name":
				_entry["value"] = container.Name
			case "State":
				_entry["value"] = container.State
				if _os.GetEnv("CONTAINER_HEALTH_STYLE") == "short" {
					_entry["representation"] = shortStateTranslations[container.State]
				} else if _os.GetEnv("CONTAINER_HEALTH_STYLE") == "icon" {
					_entry["representation"] = string(iconStateTranslations[container.State])
				}
			case "ExitCode":
				if container.ExitCode != 0 {
					_entry["value"] = fmt.Sprintf("(%d)", container.ExitCode)
				} else {
					_entry["value"] = ""
				}
			case "Image":
				_entry["value"] = container.Image
			case "Created":
				_entry["value"] = fmt.Sprintf("%d", container.Created)
				_entry["representation"] = time.Unix(container.Created, 0).Format("2006-01-02")
			}

			flat = append(flat, _entry)
		}
		row["_representation"] = flat
		rows = append(rows, row)
	}

	return rows
}

// Remove the Docker container
func (c Container) Remove(client *client.Client, force bool, removeVolumes bool) error {
	return client.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{Force: force, RemoveVolumes: removeVolumes})
}

// Pause the Docker container
func (c Container) Pause(client *client.Client) error {
	return client.ContainerPause(context.Background(), c.ID)
}

// Unpause the Docker container
func (c Container) Unpause(client *client.Client) error {
	return client.ContainerUnpause(context.Background(), c.ID)
}

// Stop the Docker container
func (c Container) Stop(client *client.Client) error {
	return client.ContainerStop(context.Background(), c.ID, container.StopOptions{})
}

// Restart the Docker container
func (c Container) Restart(client *client.Client) error {
	return client.ContainerRestart(context.Background(), c.ID, container.StopOptions{})
}

// Inspect the Docker container
func (c Container) Inspect(client *client.Client) (types.ContainerJSON, error) {
	return client.ContainerInspect(context.Background(), c.ID)
}

// Open a shell inside the Docker container
func (c Container) Shell(client *client.Client, tty *tty.TTY, channelErrors chan error, channelUpdates chan string) {
	cmd := _os.GetEnv("TTY_SERVER_COMMAND")

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          strings.Split(cmd, " "),
	}

	exec, err := client.ContainerExecCreate(context.Background(), c.ID, execConfig)
	if err != nil {
		channelErrors <- err
		return
	}

	process, err := client.ContainerExecAttach(context.Background(), exec.ID, types.ExecStartCheck{Tty: true})
	if err != nil {
		channelErrors <- err
	}
	defer process.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(tty.Stdout, process.Reader)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(process.Conn, tty.Stdin)
	}()

	channelUpdates <- "started"
	wg.Wait()
	channelUpdates <- "exited"
}

// Retrieve the public URL to access the Docker container
func (c Container) GetBrowserUrl(client *client.Client) (string, error) {
	if len(c.Ports) == 0 {
		return "", fmt.Errorf("No port is exposed on this container")
	}

	host := c.Ports[0].IP
	if host == "0.0.0.0" {
		host = "localhost"
	}

	address := ""

	for _, p := range c.Ports {
		if p.PublicPort == 443 {
			address = fmt.Sprintf("https://%s", host)
			return address, nil
		}

		if p.PublicPort == 80 {
			address = fmt.Sprintf("http://%s", host)
			return address, nil
		}
	}

	if len(address) == 0 {
		address = fmt.Sprintf("http://%s:%d", host, c.Ports[0].PublicPort)
	}

	return address, nil
}

// Rename the Docker container
func (c Container) Rename(client *client.Client, newName string) error {
	err := client.ContainerRename(context.Background(), c.ID, newName)
	return err
}

// Inspector - Retrieve the logs written by the Docker container
func (c Container) GetLogs(client *client.Client, writer io.Writer, showTimestamps bool) (*io.ReadCloser, error) {
	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: showTimestamps,
		Tail:       _os.GetEnv("CONTAINER_LOGS_TAIL"),
		Since:      _os.GetEnv("CONTAINER_LOGS_SINCE"),
		Follow:     true,
	}

	reader, err := client.ContainerLogs(context.Background(), c.ID, opts)
	if err != nil {
		return nil, err
	}

	go stdcopy.StdCopy(writer, writer, reader)

	return &reader, nil
}

// Inspector - Retrieve the full configuration of the Docker Container
func (c Container) GetConfig(client *client.Client) (ui.InspectorContent, error) {
	information, err := client.ContainerInspect(context.Background(), c.ID)

	if err != nil {
		return nil, err
	}

	// Build the first part of the config (main information)
	firstPart := ui.InspectorContentPart{Type: "rows"}
	rows := make(ui.Rows, 0)
	fields := []string{"ID", "Name", "Image", "Command"}
	for _, field := range fields {
		row := make(ui.Row)
		switch field {
		case "ID":
			row["ID"] = c.ID
			row["_representation"] = []string{"ID:", c.ID}
		case "Name":
			row["Name"] = c.Name
			row["_representation"] = []string{"Name:", c.Name}
		case "Image":
			row["Image"] = c.Name
			row["_representation"] = []string{"Image:", c.Image}
		case "Command":
			row["Command"] = strings.Join(information.Config.Entrypoint, " ") + " " + strings.Join(information.Config.Cmd, " ")
			row["_representation"] = []string{"Command:", row["Command"].(string)}
		}

		rows = append(rows, row)
	}
	firstPart.Content = rows

	separator := ui.InspectorContentPart{Type: "lines", Content: []string{"Full details:", "&nbsp;"}}

	// Build the full config using : First part // Separator // JSONBase // Mounts // Config // NetworkSettings
	allConfig := ui.InspectorContent{
		firstPart,
		separator,
		ui.InspectorContentPart{Type: "json", Content: information.ContainerJSONBase},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Mounts": information.Mounts}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Config": information.Config}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"NetworkSettings": information.NetworkSettings}},
	}

	return allConfig, nil
}

// Inspector - Retrieve the environment variables used to run the Docker container
func (c Container) GetEnv(client *client.Client) (ui.Rows, error) {
	information, err := client.ContainerInspect(context.Background(), c.ID)

	if err != nil {
		return nil, err
	}

	var env = make(ui.Rows, 0)

	raw := information.Config.Env
	for i := 0; i < len(raw); i++ {
		structured := make(ui.Row)

		pair := strings.Split(raw[i], "=")
		key := pair[0]
		value := pair[1]

		structured[key] = value
		structured["_representation"] = []string{key + ":", value}
		env = append(env, structured)
	}

	return env, nil
}

// Inspector - Retrieve the list of running processes inside the Docker container
func (c Container) GetTop(client *client.Client) (ui.Table, error) {
	if c.State == "exited" {
		return ui.Table{Headers: []string{"Notice"}, Rows: [][]string{[]string{"The container isn't running"}}}, nil
	}

	information, err := client.ContainerTop(context.Background(), c.ID, []string{})

	if err != nil {
		return ui.Table{}, err
	}

	table := ui.Table{}
	table.Headers = information.Titles
	table.Rows = information.Processes

	return table, nil
}

// Inspector - Retrieve the stats of the Docker container
func (c Container) GetStats(client *client.Client) (ui.InspectorContent, error) {
	if c.State == "exited" {
		return ui.InspectorContent{
			ui.InspectorContentPart{
				Type:    "table",
				Content: ui.Table{Headers: []string{"Notice"}, Rows: [][]string{[]string{"The container isn't running"}}},
			},
		}, nil
	}

	information, err := client.ContainerStatsOneShot(context.Background(), c.ID)
	if err != nil {
		return nil, err
	}
	defer information.Body.Close()

	var statsResult types.StatsJSON
	if err := json.NewDecoder(information.Body).Decode(&statsResult); err != nil {
		return nil, err
	}

	mainStats := ui.InspectorContentPart{Type: "rows"}
	rows := make(ui.Rows, 0)
	fields := []string{"CPU", "Memory", "Network", "PIDs"}
	for _, field := range fields {
		row := make(ui.Row)
		switch field {
		case "CPU":
			cpuUsageDelta := statsResult.CPUStats.CPUUsage.TotalUsage - statsResult.PreCPUStats.CPUUsage.TotalUsage
			cpuTotalUsageDelta := statsResult.CPUStats.SystemUsage - statsResult.PreCPUStats.SystemUsage
			value := float64(cpuUsageDelta*100) / float64(cpuTotalUsageDelta)

			row["CPU"] = value
			row["_representation"] = []string{"CPU:", fmt.Sprintf("%.2f%%", row["CPU"])}
		case "Memory":
			row["Memory"] = float64(statsResult.MemoryStats.Usage*100) / float64(statsResult.MemoryStats.Limit)
			row["_representation"] = []string{"Memory:", fmt.Sprintf("%.2f%%", row["Memory"])}
		case "Network":
			row["Network"] = fmt.Sprintf("%s / %s (RX/TX)", ui.UByteCount(statsResult.Networks["eth0"].RxBytes), ui.UByteCount(statsResult.Networks["eth0"].TxBytes))
			row["_representation"] = []string{"Network:", row["Network"].(string)}
		case "PIDs":
			row["PIDs"] = statsResult.PidsStats.Current
			row["_representation"] = []string{"PIDs:", strconv.FormatUint(statsResult.PidsStats.Current, 10)}
		}

		rows = append(rows, row)
	}
	mainStats.Content = rows

	separator := ui.InspectorContentPart{Type: "lines", Content: []string{"Full stats:", "&nbsp;"}}

	return ui.InspectorContent{
		mainStats,
		separator,
		ui.InspectorContentPart{Type: "json", Content: statsResult},
	}, nil

}
