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

	ServerError ServerMessageType = 0
	RoomUpdate  ServerMessageType = 1
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

// incoming message from client which modifies room state, nil field means no-op
type RoomActionMessage struct {
	// attached by hub
	from *Player

	// which team to join, 0 is blue 1 is red, nil means no action
	Join *int `json:"join"`

	// which option in the trivia to guess
	Guess *string `json:"guess"`

	// new chat message
	Chat *string `json:"chat"`
}

type ErrorWithMessage struct {
	message string
}

func ServerErrorHelper(msg string) OutgoingMessage {
	tobyte, _ := json.Marshal(ErrorWithMessage{msg})
	return OutgoingMessage{
		Type:    ServerError,
		Content: tobyte,
	}
}

type RoomUpdateMessage struct {
	Code    string   `json:"code"`
	Players []string `json:"players"`
	Chat    []string `json:"chat"`
}

type TriviaStateUpdateMessage struct {
}

type TriviaGameActionMessage struct {
}
