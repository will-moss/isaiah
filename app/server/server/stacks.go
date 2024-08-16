package server

import (
	"fmt"
	"io"
	"strings"
	_io "will-moss/isaiah/server/_internal/io"
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/_internal/process"
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/resources"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Placeholder used for internal organization
type Stacks struct{}

func (Stacks) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	if _os.GetEnv("DOCKER_RUNNING") == "TRUE" {
		server.SendNotification(
			session,
			ui.NotificationError(ui.NP{
				Content: ui.JSON{
					"Message": "It seems that you're running Isaiah inside a Docker container." +
						" In this case, managing stacks is unavailable because" +
						" Isaiah is bound to its container and it can't run commands on your hosting system.",
				},
			}),
		)
		return
	}

	switch command.Action {

	// Single - Default menu
	case "stack.menu":
		actions := resources.StackSingleActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - Bulk menu
	case "stacks.bulk":
		actions := resources.StacksBulkActions()
		server.SendNotification(session, ui.NotificationData(ui.NP{Content: ui.JSON{"Actions": actions}}))

	// Bulk - List
	case "stacks.list":
		columns := strings.Split(_os.GetEnv("COLUMNS_STACKS"), ",")
		stacks := resources.StacksList(server.Docker)

		rows := stacks.ToRows(columns)
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Tab": ui.Tab{Key: "stacks", Title: "Stacks", Rows: rows, SortBy: _os.GetEnv("SORTBY_STACKS")}},
			}),
		)

	// Bulk - Update
	case "stacks.update":
		stacks := resources.StacksList(server.Docker)

		hasErrored := false
		for _, stack := range stacks {
			server.SendNotification(
				session,
				ui.NotificationInfo(ui.NP{
					Content: ui.JSON{"Message": fmt.Sprintf("Updating stack %s...", stack.Name)},
				}),
			)

			err := stack.Update(server.Docker)

			if err != nil {
				server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
				hasErrored = true
				break
			}
		}

		if !hasErrored {
			server.SendNotification(
				session,
				ui.NotificationSuccess(ui.NP{
					Content: ui.JSON{"Message": "Your stacks were all succesfully updated"},
					Follow:  "init",
				}))
		}

	// Single - Up
	case "stack.up":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)

		if strings.HasPrefix(stack.Status, "running") {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": "Your stack is already up and running"}}))
			break
		}

		err := stack.Up(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your stack was succesfully started"},
				Follow:  "init",
			}))

	// Single - Pause/Unpause
	case "stack.pause":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)

		var err error
		var newState string
		if strings.HasPrefix(stack.Status, "paused") {
			err = stack.Unpause(server.Docker)
			newState = "unpaused"
		} else {
			err = stack.Pause(server.Docker)
			newState = "paused"
		}

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}
		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": fmt.Sprintf("The stack was succesfully %s", newState)}, Follow: "stacks.list",
			}),
		)

	// Single - Down
	case "stack.down":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)

		if !strings.HasPrefix(stack.Status, "running") {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": "Your stack isn't running"}}))
			break
		}

		err := stack.Down(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your stack was succesfully stopped"},
				Follow:  "init",
			}))

	// Single - Update
	case "stack.update":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)
		err := stack.Update(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your stack was succesfully updated"},
				Follow:  "stacks.list",
			}))

	// Single - Restart
	case "stack.restart":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)
		err := stack.Restart(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{
				Content: ui.JSON{"Message": "Your stack was succesfully restarted"},
				Follow:  "stacks.list",
			}))

	// Single - Create
	case "stack.create":
		task := process.LongTask{
			Function: resources.StackCreate,
			Args:     command.Args, // Expects : { "Content": <string> }
			OnStep: func(update string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{
						Content: ui.JSON{
							"Message": update,
						},
					}),
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
						Content: ui.JSON{"Message": "The stack was succesfully created"}, Follow: "init",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Single - Retrieve configuration for editing it client-side
	case "stack.edit.prepare":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)
		config, err := stack.GetRawConfig(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationPrompt(ui.NP{
				Content: ui.JSON{
					"RunLocalCommand": true,
					"Input": ui.JSON{
						"Name":         "Edit stack configuration",
						"DefaultValue": config,
						"Type":         "textarea",
						"Placeholder":  "Please fill in the content of your updated docker-compose.yml file",
					},
					"Command": "_editStack",
				},
			}),
		)

	// Single - Edit a stack (down, overwrite, up)
	case "stack.edit":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)

		task := process.LongTask{
			Function: stack.Edit,
			Args:     command.Args, // Expects : { "Content": <string> }
			OnStep: func(update string) {
				server.SendNotification(
					session,
					ui.NotificationInfo(ui.NP{
						Content: ui.JSON{
							"Message": update,
						},
					}),
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
						Content: ui.JSON{"Message": "Your stack was succesfully edited (down, overwrite, up)"}, Follow: "init",
					}),
				)
			},
		}
		task.RunSync(server.Docker)

	// Single - Get inspector tabs
	case "stack.inspect.tabs":
		tabs := resources.StacksInspectorTabs()
		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{"Inspector": ui.JSON{"Tabs": tabs}},
			}),
		)

	// Single - Inspect services (containers)
	case "stack.inspect.services":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)
		services, err := stack.GetServices(server.Docker)

		if err != nil {
			server.SendNotification(session, ui.NotificationError(ui.NP{Content: ui.JSON{"Message": err.Error()}}))
			break
		}

		server.SendNotification(
			session,
			ui.NotificationData(ui.NP{
				Content: ui.JSON{
					"Inspector": ui.JSON{
						"Content": services,
					},
				},
			}),
		)

	// Single - Inspect full configuration
	case "stack.inspect.config":
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)
		config, err := stack.GetConfig(server.Docker)

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

	// Single - Inspect logs
	case "stack.inspect.logs":
		var showTimestamps = command.Args["showTimestamps"].(bool)
		var stack resources.Stack
		mapstructure.Decode(command.Args["Resource"], &stack)

		stream, err := stack.GetLogs(
			server.Docker,
			_io.CustomWriter{WriteFunction: func(p []byte) {
				server.SendNotification(
					session,
					ui.NotificationData(ui.NP{
						Content: ui.JSON{
							"Inspector": ui.JSON{
								"Content": ui.InspectorContent{
									ui.InspectorContentPart{Type: "lines", Content: strings.Split(string(p), "\n")},
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
