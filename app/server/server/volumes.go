package server

import (
	"fmt"
	"runtime"
	"strings"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	_session "will-moss/isaiah/server/_internal/session"
	_slices "will-moss/isaiah/server/_internal/slices"
	_strconv "will-moss/isaiah/server/_internal/strconv"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Volumes struct{}

func (Volumes) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Single - Default menu
	case "volume.menu":
		actions := resources.VolumeSingleActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Single - Remove menu
	case "volume.menu.remove":
		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)

		actions := resources.VolumeRemoveActions(volume)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "volumes.bulk":
		actions := resources.VolumesBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "volumes.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_VOLUMES"), ",")
		volumes := resources.VolumesList(server.Docker)

		rows := volumes.ToRows(columns)

		// Default communication method - Send all at once
		if _os.GetEnv("SERVER_CHUNKED_COMMUNICATION_ENABLED") != "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationData(ui.NP{
					Content: ui.JSON{"Tab": ui.Tab{Key: "volumes", Title: "Volumes", Rows: rows, SortBy: _os.GetEnv("SORTBY_VOLUMES")}}}),
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
							"Tab":        ui.Tab{Key: "volumes", Title: "Volumes", Rows: c, SortBy: _os.GetEnv("SORTBY_VOLUMES")},
							"ChunkIndex": chunkIndex,
						}}),
				)
				chunkIndex += 1
			}
		}

	// Bulk - Prune
	case "volumes.prune":
		err := resources.VolumesPrune(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the unused volumes were pruned"}, Follow: "volumes.list",
			}),
		)

	// Single - Default remove
	case "volume.remove.default":
		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)

		err := volume.Remove(server.Docker, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The volume was succesfully removed"}, Follow: "volumes.list",
			}),
		)

	// Single - Forced remove
	case "volume.remove.force":
		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)

		err := volume.Remove(server.Docker, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The volume was succesfully removed"}, Follow: "volumes.list",
			}),
		)

	// Single - Browse
	case "volume.browse":
		if runtime.GOOS == "darwin" {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Docker on MacOS. On this operating system" +
							" Docker works inside a virtual machine, and therefore volumes can't be accessed" +
							" directly."},
				}),
			)
			break
		}

		if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
			server.SendNotification(
				session,
				ui.NotificationError(ui.NP{
					Content: ui.JSON{
						"Message": "It seems that you're running Isaiah inside a Docker container." +
							" In this case, external volumes can't be accessed directly because" +
							" Isaiah is bound to its container and it can't access the volumes on your hosting system.",
					},
				}),
			)
			break
		}

		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)

		terminal := tty.New(&_io.CustomWriter{WriteFunction: func(p []byte) {
			server.SendNotification(
				session,
				ui.NotificationTty(ui.NP{Content: ui.JSON{"Output": string(p)}}),
			)
		}})
		session.Set("tty", &terminal)

		errs, updates, finished := make(chan error), make(chan string), false
		go _os.OpenShell(&terminal, errs, updates)
		go terminal.RunCommand("cd " + volume.MountPoint + "\n")

		for {
			if finished {
				break
			}

			select {
			case e := <-errs:
				server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": e.Error()}}))
			case u := <-updates:
				server.SendNotification(session, ui.NotificationTty(ui.NP{Content: ui.JSON{"Status": u, "Type": "volume"}}))
				finished = u == "exited"
			}
		}

	// Single - Get inspector tabs
	case "volume.inspect.tabs":
		tabs := resources.VolumesInspectorTabs()
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": tabs}},
			}),
		)

	// Single - Inspect full configuration
	case "volume.inspect.config":
		var volume resources.Volume
		mapstructure.Decode(command.Args["Resource"], &volume)
		config, err := volume.GetConfig(server.Docker)

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
