package main

import (
	"encoding/json"
)

type PlayerMessageType int
type ServerMessageType int

const (
	// incoming message types
	Connect    PlayerMessageType = 0
	JoinRoom   PlayerMessageType = 1
	CreateRoom PlayerMessageType = 2
	RoomAction PlayerMessageType = 3
	GameAction PlayerMessageType = 4

	// outgoing message types

	ServerError      ServerMessageType = 0
	RoomUpdate       ServerMessageType = 1
	TriviaGameUpdate ServerMessageType = 2
)

// raw from clients
type IncomingMessage struct {
	from    *Player
	Type    PlayerMessageType `json:"type"`
	Content []byte            `json:"content"`
}

// raw outgoing
type OutgoingMessage struct {
	Type    ServerMessageType `json:"type"`
	Content []byte            `json:"content"`
}

// incoming
type JoinRoomMessage struct {
	Code string `json:"code"`
}

// outgoing
type JoinRoomSuccessMessage struct {
	// player number within the room
	You int `json:"you"`
}

type ActionMessage struct {
	// attached by hub
	from *Player
}

// incoming message from client which modifies room state, nil field means no-op
type RoomActionMessage struct {
	ActionMessage

	// new chat message
	Chat *string `json:"chat"`

	// should try to start the game?
	Start *bool `json:"start"`

	// join the room? gets rerouted from JoinRoomMessage
	Join *bool `json:"join"`

	// makes the sender leave the room
	Leave *bool `json:"leave"`
}

// outgoing
type ErrorWithMessage struct {
	message string
}

// outgoing
type RoomUpdateMessage struct {
	// was the room created on this update? used to assign player on frontend as owner
	Created *bool `json:"created"`

	// room code
	Code string `json:"code"`

	// playerlist TODO make player id/name and make this optional
	Players []string `json:"players"`

	// chat logs TODO make this a delta, not entire logs
	Chat []string `json:"chat"`
}

// outgoing
type TriviaStateUpdateMessage struct {
	// list of blue team players
	BlueTeam *[]string `json:"blueTeam"`

	// list of red team players
	RedTeam *[]string `json:"redTeam"`

	// limbo (0), round(1), lobby(2)
	State int `json:"state"`

	// round time, send at start
	RoundTime *int `json:"roundTime"`

	// limbo time, sent at start
	LimboTime *int `json:"limboTime"`

	// rounds since game started
	Round int `json:"round"`
}

// incoming
type TriviaGameActionMessage struct {
	ActionMessage

	// which team to join, 0 is blue 1 is red, nil means no action
	Join *int `json:"join"`

	// which option in the trivia to guess
	Guess *string `json:"guess"`
}

// signals for internal messaging between goroutines
type InternalSignal int64

// alert Room that Trivia round timer went off
const TriviaGameTimerAlert InternalSignal = 0

// generate a server error message
func serverErrorHelper(msg string) OutgoingMessage {
	tobyte, _ := json.Marshal(ErrorWithMessage{msg})
	return OutgoingMessage{
		Type:    ServerError,
		Content: tobyte,
	}
}
