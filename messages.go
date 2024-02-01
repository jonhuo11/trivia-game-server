package main

import (
	"encoding/json"
)

type PlayerMessageType int
type ServerMessageType int

const (
	Connect    PlayerMessageType = 0
	JoinRoom   PlayerMessageType = 1
	CreateRoom PlayerMessageType = 2
	RoomAction PlayerMessageType = 3
	GameAction PlayerMessageType = 4

	ServerError      ServerMessageType = 0
	RoomUpdate       ServerMessageType = 1
	TriviaGameUpdate ServerMessageType = 2
)

type IncomingMessage struct {
	From    *Player
	Type    PlayerMessageType `json:"type"`
	Content []byte            `json:"content"`
}

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
}

type ErrorWithMessage struct {
	message string
}

func serverErrorHelper(msg string) OutgoingMessage {
	tobyte, _ := json.Marshal(ErrorWithMessage{msg})
	return OutgoingMessage{
		Type:    ServerError,
		Content: tobyte,
	}
}

// outgoing
type RoomUpdateMessage struct {
	// was the room created on this update? used to assign player on frontend as owner
	Created *bool `json:"created"`

	// room code
	Code string `json:"code"`

	// playerlist TODO make player id/name
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

	// limbo (0) or round (1)
	RoundState int `json:"roundState"`

	// round time remaining, 0 means round ended and await next update packet
	RoundTimeLeft int `json:"roundTimeLeft"`

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
