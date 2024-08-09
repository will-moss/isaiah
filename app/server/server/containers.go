package server

import (
	"fmt"
	"io"
	"strings"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/_internal/tty"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Containers struct{}

func (Containers) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Single - Default menu
	case "container.menu":
		actions := resources.ContainerSingleActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Single - Remove menu
	case "container.menu.remove":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		actions := resources.ContainerRemoveActions(container)
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "containers.bulk":
		actions := resources.ContainersBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "containers.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_CONTAINERS"), ",")
		containers := resources.ContainersList(server.Docker)

		rows := containers.ToRows(columns)
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Tab": ui.Tab{Key: "containers", Title: "Containers", Rows: rows, SortBy: _os.GetEnv("SORTBY_CONTAINERS")}}}),
		)

	// Bulk - Prune
	case "containers.prune":
		err := resources.ContainersPrune(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the unused containers were pruned"}, Follow: "containers.list",
			}),
		)

	// Bulk - Stop
	case "containers.stop":
		task := process.LongTask{
			Function: resources.ContainersStop,
			OnStep: func(id string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{Content: ui.JSON{"Message": fmt.Sprintf("Container %s was stopped", id)}}),
				)
				server.SendNotification(
					session,
					ui.NotificationLoading(),
				)
			},
			OnError: func(err error) {
				server.SendNotification(
					session,
					ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}),
				)
			},
			OnDone: func() {
				server.SendNotification(
					session,
					ui.NotificationSuccess(ui.NP{
						Content: ui.JSON{"Message": "All the containers were stopped"}, Follow: "containers.list",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Bulk - Remove
	case "containers.remove":
		err := resources.ContainersRemove(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "All the containers were removed"}, Follow: "containers.list",
			}),
		)

	// Single - Pause/Unpause
	case "container.pause":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		var newState string
		if information.State.Paused {
			err = container.Unpause(server.Docker)
			newState = "unpaused"
		} else {
			err = container.Pause(server.Docker)
			newState = "paused"
		}

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": fmt.Sprintf("The container was succesfully %s", newState)}, Follow: "containers.list",
			}),
		)

	// Single - Stop
	case "container.stop":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Stop(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully stopped"}, Follow: "containers.list",
			}),
		)

	// Single - Restart
	case "container.restart":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Restart(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully restarted"}, Follow: "containers.list",
			}),
		)

	// Single - Default remove
	case "container.remove.default":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		if information.State.Running {
			server.SendNotification(
				session,
				ui.NotificationPrompt(ui.NP{
					Content: ui.JSON{
						"Message": "You cannot remove a container unless you force it. Do you want to force it?",
						"Command": "container.remove.force",
					},
				}),
			)
			break
		}

		err = container.Remove(server.Docker, false, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Force remove
	case "container.remove.force":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Remove(server.Docker, true, false)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Default remove with volumes
	case "container.remove.default.volumes":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		information, err := container.Inspect(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		if information.State.Running {
			server.SendNotification(
				session,
				ui.NotificationPrompt(ui.NP{
					Content: ui.JSON{
						"Message": "You cannot remove a container unless you force it. Do you want to force it?",
						"Command": "container.remove.force.volumes",
					},
				}),
			)
			break
		}

		err = container.Remove(server.Docker, false, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Force remove with volumes
	case "container.remove.force.volumes":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		err := container.Remove(server.Docker, true, true)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "The container was succesfully removed"}, Follow: "containers.list",
			}),
		)

	// Single - Open shell inside container
	case "container.shell":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		terminal := tty.New(&_io.CustomWriter{WriteFunction: func(p []byte) {
			server.SendNotification(
				session,
				ui.NotificationTty(ui.NP{Content: ui.JSON{"Output": string(p)}}),
			)
		}})
		session.Set("tty", &terminal)

		go func() {
			errs, updates, finished := make(chan error), make(chan string), false
			go container.Shell(server.Docker, &terminal, errs, updates)

			for {
				if finished {
					break
				}

				select {
				case e := <-errs:
					server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": e.Error()}}))
				case u := <-updates:
					server.SendNotification(session, ui.NotificationTty(ui.NP{Content: ui.JSON{"Status": u, "Type": "container"}}))
					finished = u == "exited"
				}

			}
		}()

	// Single - Open in browser
	case "container.browser":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		address, err := container.GetBrowserUrl(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Address": address}}))

	// Single - Rename
	case "container.rename":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		err := container.Rename(server.Docker, command.Args["Name"].(string))

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your container was succesfully renamed"},
				Follow:  "containers.list",
			}))

	// Single - Update
	case "container.update":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		err := container.Update(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your container was succesfully updated"},
				Follow:  "containers.list",
			}))

	// Single - Get inspector tabs
	case "container.inspect.tabs":
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": resources.ContainersInspectorTabs()}},
			}),
		)

	// Single - Inspect logs
	case "container.inspect.logs":
		var showTimestamps = command.Args["showTimestamps"].(bool)
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		stream, err := container.GetLogs(
			server.Docker,
			_io.CustomWriter{WriteFunction: func(p []byte) {
				server.SendNotification(
					session,
					ui.NotificationData(ui.NP{
						Content: ui.JSON{
							"Inspector": ui.JSON{
								"Content": ui.InspectorContent{
									ui.InspectorContentPart{Type: "lines", Content: []string{string(p)}},
								},
							},
						},
					}),
				)
			}},
			showTimestamps,
		)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			if _stream, exists := session.Get("stream"); exists {
				(*_stream.(*io.ReadCloser)).Close()
				session.UnSet("stream")
			}
			break
		}

		session.Set("stream", stream)

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{Content: ui.JSON{}}),
		)

	// Single - Inspect full configuration
	case "container.inspect.config":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		config, err := container.GetConfig(server.Docker)

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

	// Single - Inspect top (running processes)
	case "container.inspect.top":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)

		processes, err := container.GetTop(server.Docker)
		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": ui.InspectorContent{
							ui.InspectorContentPart{Type: "table", Content: processes},
						},
					},
				},
			}),
		)

	// Single - Inspect environment variables
	case "container.inspect.env":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		env, err := container.GetEnv(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": ui.InspectorContent{
							ui.InspectorContentPart{Type: "rows", Content: env},
						},
					},
				},
			}),
		)

	// Single - Inspect stats
	case "container.inspect.stats":
		var container resources.Container
		mapstructure.Decode(command.Args["Resource"], &container)
		stats, err := container.GetStats(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": stats,
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
