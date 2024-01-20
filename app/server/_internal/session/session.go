package session

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Represent a Generic Session entity
// The only reason this type exists is to provide inheritance
// In some parts of the code, a melody.Session will be a GenericSession
// In other parts of the code, a Session will be a GenericSession
// GenericSession hence enables us to use both types transparently
type GenericSession interface {
	Set(string, interface{})
	Get(string) (interface{}, bool)
	UnSet(key string)
	Write([]byte) error
}

// Stripped down copy of melody.Session
// This version is used in place of melody.Session when current node is an agent
// We can't use melody.Session as an agent because this requires a server, yet the agent
// isn't a server. It is a client connecting to the master node
type Session struct {
	Connection *websocket.Conn
	Keys       map[string]interface{}
	rwmutex    sync.RWMutex
	mutex      sync.Mutex
}

func Create(connection *websocket.Conn) *Session {
	return &Session{Connection: connection}
}

func (s *Session) Write(msg []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.Connection.WriteMessage(websocket.TextMessage, msg)
}

// Custom reimplementation of melody.Session.Set that first checks if an "initiator"
// field is set, and sets the value associated with <initiator_id>_<key> field if it is
// Otherwise, simply sets the value associated with <key>
func (s *Session) Set(key string, value interface{}) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()

	if s.Keys == nil {
		s.Keys = make(map[string]interface{})
	}

	if key == "initiator" {
		s.Keys[key] = value
		return
	}

	if initiator, ok := s.Keys["initiator"]; ok {
		s.Keys[initiator.(string)+"_"+key] = value
		return
	}

	s.Keys[key] = value
}

// Same custom mechanism as Set (retrieve value associated with <initiator_id>_<key> when applicable)
func (s *Session) Get(key string) (value interface{}, exists bool) {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	if s.Keys != nil {
		if key == "initiator" {
			value, exists := s.Keys[key]
			return value, exists
		}

		if initiator, ok := s.Keys["initiator"]; ok {
			value, exists := s.Keys[initiator.(string)+"_"+key]
			return value, exists
		}

		value, exists := s.Keys[key]
		return value, exists
	}

	return nil, false
}

func (s *Session) UnSet(key string) {
	s.rwmutex.Lock()
	defer s.rwmutex.Unlock()

	if s.Keys != nil {
		if key == "initiator" {
			delete(s.Keys, key)
			return
		}

		if initiator, ok := s.Keys["initiator"]; ok {
			delete(s.Keys, initiator.(string)+"_"+key)
			return
		}

		delete(s.Keys, key)
	}
}
