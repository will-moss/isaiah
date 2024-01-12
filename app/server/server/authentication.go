package server

import (
	_os "will-moss/isaiah/server/_internal/os"
	"will-moss/isaiah/server/ui"

	"github.com/olahol/melody"
)

type Authentication struct{}

func (Authentication) RunCommand(server *Server, session *melody.Session, command ui.Command) {
	switch command.Action {

	// Command : Authenticate the client by password
	case "auth.login":
		if _os.GetEnv("AUTHENTICATION_ENABLED") != "TRUE" {
			session.Set("authenticated", true)
			server.SendNotification(session, ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "Your are now authenticated",
					},
				},
			}),
			)
			break
		}

		password := command.Args["Password"]

		if password != _os.GetEnv("AUTHENTICATION_SECRET") {
			session.Set("authenticated", false)
			server.SendNotification(
				session,
				ui.NotificationAuth(ui.NP{
					Type: ui.TypeError,
					Content: ui.JSON{
						"Authentication": ui.JSON{
							"Message": "Invalid password",
						},
					},
				}),
			)
			break
		}

		session.Set("authenticated", true)

		server.SendNotification(
			session,
			ui.NotificationAuth(ui.NP{
				Type: ui.TypeSuccess,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "Your are now authenticated",
					},
				},
			}),
		)

	// Command : Log out the client
	case "auth.logout":
		session.Set("authenticated", false)

	// Command not found
	default:
		server.SendNotification(
			session,
			ui.NotificationAuth(ui.NP{
				Type: ui.TypeError,
				Content: ui.JSON{
					"Authentication": ui.JSON{
						"Message": "You are not authenticated yet",
					},
				},
			}),
		)
	}
}
