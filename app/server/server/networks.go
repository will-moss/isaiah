package server

import (
	"fmt"
	"strings"
	_os "will-moss/isaiah/server/_internal/os"
	_session "will-moss/isaiah/server/_internal/session"
	_slices "will-moss/isaiah/server/_internal/slices"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Networks struct{}

func (Networks) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Single - Default menu
	case "network.menu":
		var network resources.Network
		mapstructure.Decode(command.Args["Resource"], &network)

		actions := resources.NetworkSingleActions(network)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Single - Remove menu
	case "network.menu.remove":
		var network resources.Network
		mapstructure.Decode(command.Args["Resource"], &network)

		actions := resources.NetworkRemoveActions(network)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "networks.bulk":
		actions := resources.NetworksBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "networks.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_NETWORKS"), ",")
		networks := resources.NetworksList(server.Docker)

		rows := networks.ToRows(columns)

		// Default communication method - Send all at once
		if _os.GetEnv("SERVER_CHUNKED_COMMUNICATION_ENABLED") != "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationData(ui.NP{
					Content: ui.JSON{"Tab": ui.Tab{Key: "networks", Title: "Networks", Rows: rows, SortBy: _os.GetEnv("SORTBY_NETWORKS")}}}),
			)
		} else {
			// Chunked communication method, send resources chunk by chunk
			chunkSize := int(_strconv.ParseInt(_os.GetEnv("SERVER_CHUNKED_COMMUNICATION_SIZE"), 10, 64))
			chunkIndex := 1
			chunks := _slices.Chunk(rows, chunkSize)
			for _, c := range chunks {
				server.SendNotification(
					session,
					ui.NotificationDataChunk(ui.NP{
						Content: ui.JSON{
							"Tab":        ui.Tab{Key: "networks", Title: "Networks", Rows: c, SortBy: _os.GetEnv("SORTBY_NETWORKS")},
							"ChunkIndex": chunkIndex,
						}}),
				)
				chunkIndex += 1
			}
		}

	// Bulk - Prune
	case "networks.prune":
		err := resources.NetworksPrune(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the unused networks were pruned"}, Follow: "networks.list",
			}),
		)

	// Single - Default remove
	case "network.remove.default":
		var network resources.Network
		mapstructure.Decode(command.Args["Resource"], &network)

		err := network.Remove(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The network was succesfully removed"}, Follow: "networks.list",
			}),
		)

	// Single - Get inspector tabs
	case "network.inspect.tabs":
		tabs := resources.NetworksInspectorTabs()
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": tabs}},
			}),
		)

	// Single - Inspect full configuration
	case "network.inspect.config":
		var network resources.Network
		mapstructure.Decode(command.Args["Resource"], &network)
		config, err := network.GetConfig(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": config,
					},
				},
			}),
		)

	// Command not found
	default:
		server.SendNotification(
			session,
			ui.NotificationError(ui.NP{
				Content: ui.JSON{
					"Message": fmt.Sprintf("This command is unknown, unsupported, or not implemented yet : %s", command.Action),
				},
			}),
		)
	}
}
