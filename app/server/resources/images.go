package resources

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"will-moss/isaiah/server/_internal/process"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/fatih/structs"
)

// Represent a Docker image
type Image struct {
	ID      string
	Name    string
	Version string
	Size    int64
}

// Represent an array of Docker images
type Images []Image

// Retrieve all inspector tabs for Docker images
func ImagesInspectorTabs() []string {
	return []string{"Config"}
}

// Retrieve all the single actions associated with Docker images
func ImageSingleActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:            "remove image",
			Command:          "image.menu.remove",
			Key:              "d",
			RequiresResource: true,
		},
		ui.MenuAction{
			Label:            "run image",
			Command:          "run_restart",
			Key:              "r",
			RequiresResource: false,
			RunLocally:       true,
		},
		ui.MenuAction{
			Label:            "open on Docker Hub",
			Command:          "hub",
			Key:              "h",
			RequiresResource: false,
			RunLocally:       true,
		},
		ui.MenuAction{
			Label:            "pull a new image",
			Command:          "pull",
			Key:              "P",
			RequiresResource: false,
			RunLocally:       true,
		},
	)
	return actions
}

// Retrieve all the remove actions associated with Docker images
func ImageRemoveActions(v Volume) []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove",
			Label:            fmt.Sprintf("<em>docker image rm %s</em> ?", v.Name),
			Command:          "image.remove.default",
			RequiresResource: true,
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove without deleting untagged parents",
			Label:            fmt.Sprintf("<em>docker image rm --no-prune %s</em> ?", v.Name),
			Command:          "image.remove.default.unprune",
			RequiresResource: true,
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "force remove",
			Label:            fmt.Sprintf("<em>docker image rm --force %s</em> ?", v.Name),
			Command:          "image.remove.force",
			RequiresResource: true,
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "force remove without deleting untagged parents ",
			Label:            fmt.Sprintf("<em>docker image rm --no-prune --force %s</em> ?", v.Name),
			Command:          "image.remove.force.unprune",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the bulk actions associated with Docker images
func ImagesBulkActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "prune unused images",
			Prompt:  "Are you sure you want to prune all unused images?",
			Command: "images.prune",
		},
	)
	return actions
}

// Retrieve all Docker images
func ImagesList(client *client.Client) Images {
	reader, err := client.ImageList(context.Background(), types.ImageListOptions{All: true})

	if err != nil {
		return []Image{}
	}

	var images []Image
	for i := 0; i < len(reader); i++ {
		var summary = reader[i]

		var image Image
		image.ID = summary.ID

		if len(summary.RepoTags) > 0 {
			if strings.Contains(summary.RepoTags[0], ":") {
				parts := strings.Split(summary.RepoTags[0], ":")
				image.Name = parts[0]
				image.Version = parts[1]
			}
		} else {
			image.Name = "<none>"
			image.Version = "<none>"
		}

		image.Size = summary.Size
		images = append(images, image)
	}

	return images
}

// Prune unused Docker images
func ImagesPrune(client *client.Client) error {
	args := filters.NewArgs(filters.KeyValuePair{Key: "dangling", Value: "false"})
	_, err := client.ImagesPrune(context.Background(), args)

	return err
}

// Turn the list of Docker images into a list of rows representing them
func (images Images) ToRows(columns []string) ui.Rows {
	var rows = make(ui.Rows, 0)

	sort.Slice(images, func(i, j int) bool {
		if images[i].Name == "<none>" {
			return false
		}
		if images[j].Name == "<none>" {
			return true
		}
		return images[i].Name < images[j].Name
	})

	for i := 0; i < len(images); i++ {
		image := images[i]

		row := structs.Map(image)
		var flat = make([]map[string]string, 0)

		for j := 0; j < len(columns); j++ {
			_entry := make(map[string]string)
			_entry["field"] = columns[j]

			switch columns[j] {
			case "ID":
				_entry["value"] = image.ID
			case "Name":
				_entry["value"] = image.Name
			case "Version":
				_entry["value"] = image.Version
			case "Size":
				_entry["value"] = strconv.FormatInt(image.Size, 10)
				_entry["representation"] = ui.ByteCount(image.Size)
			}

			flat = append(flat, _entry)
		}
		row["_representation"] = flat
		rows = append(rows, row)
	}

	return rows
}

// Remove the Docker image
func (i Image) Remove(client *client.Client, force bool, prune bool) error {
	_, err := client.ImageRemove(context.Background(), i.ID, types.ImageRemoveOptions{Force: force, PruneChildren: prune})
	return err
}

// Pull a new Docker image
func ImagePull(c *client.Client, m process.LongTaskMonitor, args map[string]interface{}) {
	name := args["Image"].(string)
	rc, err := c.ImagePull(context.Background(), name, types.ImagePullOptions{})

	if err != nil {
		m.Errors <- err
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		scanner := bufio.NewScanner(rc)
		for scanner.Scan() {
			m.Results <- scanner.Text()
		}
		wg.Done()
	}()

	wg.Wait()
	m.Done <- true
}

// Inspector - Retrieve the full configuration associated with a Docker image
func (i Image) GetConfig(client *client.Client) (ui.InspectorContent, error) {
	information, _, err := client.ImageInspectWithRaw(context.Background(), i.ID)

	if err != nil {
		return nil, err
	}

	// Build the first part of the config (main information)
	firstPart := ui.InspectorContentPart{Type: "rows"}
	rows := make(ui.Rows, 0)
	fields := []string{"Name", "ID", "Tags", "Size", "Created"}
	for _, field := range fields {
		row := make(ui.Row)
		switch field {
		case "Name":
			row["Name"] = i.Name
			row["_representation"] = []string{"Name:", i.Name}
		case "ID":
			row["ID"] = i.ID
			row["_representation"] = []string{"ID:", i.ID}
		case "Tags":
			row["Tags"] = information.RepoTags
			row["_representation"] = []string{"Tags:", strings.Join(information.RepoTags, ", ")}
		case "Size":
			row["Size"] = information.Size
			row["_representation"] = []string{"Size:", ui.ByteCount(information.Size)}
		case "Created":
			row["Created"] = information.Created
			row["_representation"] = []string{"Created:", information.Created}
		}

		rows = append(rows, row)
	}
	firstPart.Content = rows

	separator := ui.InspectorContentPart{Type: "lines", Content: []string{"&nbsp;", "&nbsp;"}}

	// Build the image's history
	table := ui.Table{}
	table.Headers = []string{"ID", "TAG", "SIZE", "COMMAND"}

	history, err := client.ImageHistory(context.Background(), i.ID)
	if err == nil {
		rows := make([][]string, 0)
		for _, entry := range history {
			_id := "&lt;none&gt;"
			if entry.ID != "" {
				if len(entry.ID) > 17 {
					_id = entry.ID[7:17]
				}
			}

			_tag := ""
			if len(entry.Tags) > 0 {
				_tag = entry.Tags[0]
			}

			rows = append(
				rows,
				[]string{
					_id,
					_tag,
					ui.ByteCount(entry.Size),
					entry.CreatedBy,
				},
			)
		}
		table.Rows = rows
	} else {
		log.Print(err)
	}

	// Build the full config using : First part // Separator // History
	allConfig := ui.InspectorContent{
		firstPart,
		separator,
		ui.InspectorContentPart{Type: "table", Content: table},
	}

	return allConfig, nil
}

// Create and start a new Docker container based on the Docker image
func (i Image) Run(client *client.Client, name string) error {
	response, err := client.ContainerCreate(
		context.Background(),
		&container.Config{Image: i.Name},
		nil,
		nil,
		nil,
		name,
	)

	if err != nil {
		return err
	}

	// Start the container
	err = client.ContainerStart(
		context.Background(),
		response.ID,
		types.ContainerStartOptions{},
	)

	return err
}
