package resources

import (
	"context"
	"fmt"
	"sort"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"

	"github.com/fatih/structs"
)

// Represent a Docker volume
type Volume struct {
	Name       string
	Driver     string
	MountPoint string
}

// Represent an array of Docker volumes
type Volumes []Volume

// Retrieve all inspector tabs for Docker volumes
func VolumesInspectorTabs() []string {
	return []string{"Config"}
}

// Retrieve all the single actions associated with Docker volumes
func VolumeSingleActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:            "remove volume",
			Command:          "volume.menu.remove",
			Key:              "d",
			RequiresResource: true,
		},
	)

	actions = append(
		actions,
		ui.MenuAction{
			Label:            "browse volume in shell",
			Command:          "volume.browse",
			Key:              "B",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the remove actions associated with Docker volumes
func VolumeRemoveActions(v Volume) []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove",
			Label:            fmt.Sprintf("<em>docker volume rm %s</em> ?", v.Name),
			Command:          "volume.remove.default",
			RequiresResource: true,
		},
	)
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "force remove",
			Label:            fmt.Sprintf("<em>docker volume rm --force %s</em> ?", v.Name),
			Command:          "volume.remove.force",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the bulk actions associated with Docker volumes
func VolumesBulkActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "prune unused volumes",
			Prompt:  "Are you sure you want to prune all unused volumes?",
			Command: "volumes.prune",
		},
	)
	return actions
}

// Retrieve all Docker volumes
func VolumesList(client *client.Client) Volumes {
	reader, err := client.VolumeList(context.Background(), volume.ListOptions{})

	if err != nil {
		return []Volume{}
	}

	var volumes []Volume
	for i := 0; i < len(reader.Volumes); i++ {
		var information = reader.Volumes[i]

		var volume Volume
		volume.Name = information.Name
		volume.Driver = information.Driver
		volume.MountPoint = information.Mountpoint

		volumes = append(volumes, volume)
	}

	return volumes
}

// Prune unused Docker volumes
func VolumesPrune(client *client.Client) error {
	_, err := client.VolumesPrune(context.Background(), filters.Args{})
	return err
}

// Turn the list of Docker volumes into a list of rows representing them
func (volumes Volumes) ToRows(columns []string) ui.Rows {
	var rows = make(ui.Rows, 0)

	sort.Slice(volumes, func(i, j int) bool {
		return volumes[i].Name < volumes[j].Name
	})

	for i := 0; i < len(volumes); i++ {
		volume := volumes[i]

		row := structs.Map(volume)
		var flat = make([]map[string]string, 0)

		for j := 0; j < len(columns); j++ {
			_entry := make(map[string]string)
			_entry["field"] = columns[j]

			switch columns[j] {
			case "Name":
				_entry["value"] = volume.Name
			case "Driver":
				_entry["value"] = volume.Driver
			case "MountPoint":
				_entry["value"] = volume.MountPoint
			}

			flat = append(flat, _entry)
		}
		row["_representation"] = flat
		rows = append(rows, row)
	}

	return rows
}

// Remove the Docker Volume
func (v Volume) Remove(client *client.Client, force bool) error {
	err := client.VolumeRemove(context.Background(), v.Name, force)
	return err
}

// Inspector - Retrieve the full configuration associated with a Docker volume
func (v Volume) GetConfig(client *client.Client) (ui.InspectorContent, error) {
	information, err := client.VolumeInspect(context.Background(), v.Name)

	if err != nil {
		return nil, err
	}

	// Build the first part of the config (main information)
	firstPart := ui.InspectorContentPart{Type: "rows"}
	rows := make(ui.Rows, 0)
	fields := []string{"Name", "Driver", "Scope", "Mountpoint"}
	for _, field := range fields {
		row := make(ui.Row)
		switch field {
		case "Name":
			row["Name"] = v.Name
			row["_representation"] = []string{"Name:", v.Name}
		case "Driver":
			row["Driver"] = v.Driver
			row["_representation"] = []string{"Driver:", v.Driver}
		case "Scope":
			row["Scope"] = information.Scope
			row["_representation"] = []string{"Scope:", information.Scope}
		case "Mountpoint":
			row["Mountpoint"] = information.Mountpoint
			row["_representation"] = []string{"Mountpoint:", information.Mountpoint}
		}

		rows = append(rows, row)
	}
	firstPart.Content = rows

	// Build the full config using : First part // Labels // Options // Status
	allConfig := ui.InspectorContent{
		firstPart,
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Labels": information.Labels}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Options": information.Options}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Status": information.Status}},
	}

	return allConfig, nil
}
