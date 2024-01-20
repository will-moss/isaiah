package server

import (
	_session "will-moss/isaiah/server/_internal/session"
	"will-moss/isaiah/server/ui"

	"github.com/mitchellh/mapstructure"
)

// Represent an Isaiah agent
type Agent struct {
	Name string
}

// Represent an array of Isaiah agents
type AgentsArray []Agent

// Placeholder used for internal organization
type Agents struct{}

func (handler Agents) RunCommand(server *Server, session _session.GenericSession, command ui.Command) {
	switch command.Action {

	// Command : Register a new agent
	case "agent.register":
		var agent Agent
		mapstructure.Decode(command.Args["Resource"], &agent)

		for _, name := range server.Agents.ToStrings() {
			if name == agent.Name {
				server.SendNotification(
					session,
					ui.NotificationError(ui.NP{Content: ui.JSON{
						"Message": "This name is already taken. Please use another unique name for your agent",
					}}),
				)
				return
			}
		}

		session.Set("agent", agent)
		server.Agents = append(server.Agents, agent)

		server.SendNotification(
			session,
			ui.NotificationSuccess(ui.NP{Content: ui.JSON{"Message": "The agent was succesfully registered"}}),
		)

		// Notify all the clients about the new agent's registration
		notification := ui.NotificationData(ui.NotificationParams{Content: ui.JSON{"Agents": server.Agents.ToStrings()}})
		server.Melody.Broadcast(notification.ToBytes())

	// Command : Agent replies to a specific client
	case "agent.reply":
		var to string
		mapstructure.Decode(command.Args["To"], &to)

		if to == "" {
			return
		}

		sessions, _ := server.Melody.Sessions()
		for index := range sessions {
			_session := sessions[index]

			if id, exists := _session.Get("id"); !exists || id != to {
				continue
			}

			var _notification ui.Notification
			mapstructure.Decode(command.Args["Notification"], &_notification)
			_session.Write(_notification.ToBytes())
			break
		}

		// -> Agent's "logout" is performed when the websocket connection is terminated

	}

}

func (agents AgentsArray) ToStrings() []string {
	arr := make([]string, 0)

	for _, v := range agents {
		arr = append(arr, v.Name)
	}

	return arr
}
