package resources

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"will-moss/isaiah/server/ui"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/fatih/structs"
)

// Represent a Docker netowrk
type Network struct {
	ID     string
	Name   string
	Driver string
}

// Represent a array of Docker networks
type Networks []Network

// Retrieve all inspector tabs for networks
func NetworksInspectorTabs() []string {
	return []string{"Config"}
}

// Retrieve all the single actions associated with Docker networks
func NetworkSingleActions(n Network) []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "d",
			Label:            "remove network",
			Command:          "network.menu.remove",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the remove actions associated with Docker networks
func NetworkRemoveActions(n Network) []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Key:              "remove",
			Label:            fmt.Sprintf("<em>docker network rm %s</em> ?", n.Name),
			Command:          "network.remove.default",
			RequiresResource: true,
		},
	)
	return actions
}

// Retrieve all the bulk actions associated with Docker networks
func NetworksBulkActions() []ui.MenuAction {
	var actions []ui.MenuAction
	actions = append(
		actions,
		ui.MenuAction{
			Label:   "prune unused networks",
			Prompt:  "Are you sure you want to prune all unused networks?",
			Command: "networks.prune",
		},
	)
	return actions
}

// Retrieve all Docker networks
func NetworksList(client *client.Client) Networks {
	reader, err := client.NetworkList(context.Background(), network.ListOptions{})

	if err != nil {
		return []Network{}
	}

	var networks []Network
	for i := 0; i < len(reader); i++ {
		var information = reader[i]

		var network Network
		network.ID = information.ID
		network.Name = information.Name
		network.Driver = information.Driver

		networks = append(networks, network)
	}

	return networks
}

// Count the number of Docker networks
func NetworksCount(client *client.Client) int {
	images, err := client.NetworkList(context.Background(), network.ListOptions{})

	if err != nil {
		return 0
	}

	return len(images)
}

// Prune unused Docker networks
func NetworksPrune(client *client.Client) error {
	_, err := client.NetworksPrune(context.Background(), filters.Args{})
	return err
}

// Remove the Docker network
func (n Network) Remove(client *client.Client) error {
	err := client.NetworkRemove(context.Background(), n.ID)
	return err
}

// Turn the list of Docker networks into a list of string rows representing them
func (networks Networks) ToRows(columns []string) ui.Rows {
	var rows = make(ui.Rows, 0)

	sort.Slice(networks, func(i, j int) bool {
		return networks[i].Name < networks[j].Name
	})

	for i := 0; i < len(networks); i++ {
		network := networks[i]

		row := structs.Map(network)
		var flat = make([]map[string]string, 0)

		for j := 0; j < len(columns); j++ {
			_entry := make(map[string]string)
			_entry["field"] = columns[j]

			switch columns[j] {
			case "ID":
				_entry["value"] = network.ID
			case "Name":
				_entry["value"] = network.Name
			case "Driver":
				_entry["value"] = network.Driver
			}

			flat = append(flat, _entry)
		}
		row["_representation"] = flat
		rows = append(rows, row)
	}

	return rows
}

// Inspector - Retrieve the full configuration associated with a Docker network
func (n Network) GetConfig(client *client.Client) (ui.InspectorContent, error) {
	information, err := client.NetworkInspect(context.Background(), n.ID, network.InspectOptions{})

	if err != nil {
		return nil, err
	}

	// Build the first part of the config (main information)
	firstPart := ui.InspectorContentPart{Type: "rows"}
	rows := make(ui.Rows, 0)
	fields := []string{"ID", "Name", "Driver", "Scope", "EnabledIPV6", "Internal", "Attachable", "Ingress"}
	for _, field := range fields {
		row := make(ui.Row)
		switch field {
		case "ID":
			row["ID"] = n.Name
			row["_representation"] = []string{"ID:", n.ID}
		case "Name":
			row["Name"] = n.Name
			row["_representation"] = []string{"Name:", n.Name}
		case "Driver":
			row["Driver"] = n.Driver
			row["_representation"] = []string{"Driver:", n.Driver}
		case "Scope":
			row["Scope"] = information.Scope
			row["_representation"] = []string{"Scope:", information.Scope}
		case "EnabledIPV6":
			row["EnabledIPV6"] = information.EnableIPv6
			row["_representation"] = []string{"EnabledIPV6:", strconv.FormatBool(information.EnableIPv6)}
		case "Internal":
			row["Internal"] = information.Internal
			row["_representation"] = []string{"Internal:", strconv.FormatBool(information.Internal)}
		case "Attachable":
			row["Attachable"] = information.Attachable
			row["_representation"] = []string{"Attachable:", strconv.FormatBool(information.Attachable)}
		case "Ingress":
			row["Ingress"] = information.Ingress
			row["_representation"] = []string{"Ingress:", strconv.FormatBool(information.Ingress)}
		}

		rows = append(rows, row)
	}
	firstPart.Content = rows

	// Build the full config using : First part // Containers // Labels // Options
	allConfig := ui.InspectorContent{
		firstPart,
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Containers": information.Containers}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Labels": information.Labels}},
		ui.InspectorContentPart{Type: "json", Content: ui.JSON{"Options": information.Options}},
	}

	return allConfig, nil
}
