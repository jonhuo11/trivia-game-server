package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered players.
	players map[*Player]bool

	// Inbound messages from the clients.
	incoming chan IncomingMessage

	// Register requests from the clients.
	register chan *Player

	// Unregister requests from clients.
	unregister chan *Player

	rooms map[string]*Room
}

func newHub() *Hub {
	return &Hub{
		incoming:   make(chan IncomingMessage),
		register:   make(chan *Player),
		unregister: make(chan *Player),
		players:    make(map[*Player]bool),
		rooms:      make(map[string]*Room),
	}
}

func (h *Hub) joinRoom(p *Player, code string) {
	if p.room != nil {
		p.send <- ServerErrorHelper("this player is already in a room")
		return
	}
	if room, in := h.rooms[code]; !in {
		p.send <- ServerErrorHelper("this room does not exist")
		return
	} else {
		room.mu.Lock()
		room.players[p] = len(room.players)
		p.room = room
		room.mu.Unlock()
		room.broadcastRoomUpdate()
	}
}

func (h *Hub) createRoom(creator *Player) {
	if creator.room != nil {
		creator.send <- ServerErrorHelper("this player is already in a room")
		return
	}
	id := uuid.New().String()
	newroom := &Room{
		code:                  id,
		players:               map[*Player]int{creator: 0},
		chat:                  []string{fmt.Sprintf("Welcome to room %s", id)},
		state:                 Lobby,
		gamestate:             TriviaState{},
		incomingRoomActions:   make(chan RoomActionMessage),
		incomingTriviaActions: make(chan TriviaGameActionMessage),
	}
	h.rooms[id] = newroom
	creator.room = newroom

	newroom.broadcastRoomUpdate()

	go newroom.run()
}

func (h *Hub) run() {
	for {
		select {
		case player := <-h.register:
			h.players[player] = true
		case player := <-h.unregister:
			if player.room != nil {
				if _, in := h.rooms[player.room.code]; in {
					playerroom := h.rooms[player.room.code]
					playerroom.writeChat(fmt.Sprintf("Player %d left the room", playerroom.players[player]))
					playerroom.removePlayer(player)
					playerroom.broadcastRoomUpdate()
				}
			}
			delete(h.players, player)
			close(player.send)
			fmt.Println("Unregistered client and removed from room")
			break
		case message := <-h.incoming:
			switch message.Type {
			case Connect:
				fmt.Println("New player connected from ", message.From.conn.RemoteAddr())
				break
			case JoinRoom:
				m := JoinRoomMessage{}
				if err := json.Unmarshal(message.Content, &m); err != nil {
					message.From.send <- ServerErrorHelper("Bad format")
				} else {
					h.joinRoom(message.From, m.Code)
				}
				break
			case CreateRoom:
				h.createRoom(message.From)
				break
			case RoomAction:
				// RoomAction is join/leave room, switch team, send chat message

				// parse the message content as a room message and send to room handler
				if message.From.room != nil {
					rm := RoomActionMessage{from: message.From}
					if err := json.Unmarshal(message.Content, &rm); err != nil {
						message.From.send <- ServerErrorHelper("Bad RoomActionMessage format")
					} else {
						message.From.room.incomingRoomActions <- rm
					}
				} else {
					message.From.send <- ServerErrorHelper("Not in a room")
				}
				break
			default:
				break
			}
		}
	}
}
