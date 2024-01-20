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

type JoinRoomMessage struct {
	Code string `json:"code"`
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
}
