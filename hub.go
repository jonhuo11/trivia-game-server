package main

import (
	"encoding/json"

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

	// List of rooms
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
		p.send <- serverErrorHelper("this player is already in a room")
		return
	}
	if room, in := h.rooms[code]; !in {
		p.send <- serverErrorHelper("this room does not exist")
		return
	} else {
		room.join(p)
		room.broadcastRoomUpdate(false)
		room.game.broadcastGameUpdate(true)
	}
}

func (h *Hub) createRoom(creator *Player) {
	if creator.room != nil {
		creator.send <- serverErrorHelper("this player is already in a room")
		return
	}
	id := uuid.New().String()
	newroom := newRoom(false)
	newroom.join(creator)
	h.rooms[id] = newroom
	creator.room = newroom

	newroom.broadcastRoomUpdate(true)

	go func() { // TODO add stopper
		for {
			newroom.run()
		}
	}()
}

func (h *Hub) run() {
	for {
		select {
		case player := <-h.register:
			h.players[player] = true
		case player := <-h.unregister:
			if player.room != nil {
				if _, in := h.rooms[player.room.code]; in {
					ram := RoomActionMessage{}
					ram.from = player
					t := true
					ram.Leave = &t
					h.rooms[player.room.code].incomingRoomActions <- ram
				}
			}
			delete(h.players, player)
			close(player.send)
			//fmt.Println("Unregistered client and removed from room")
			break
		case message := <-h.incoming:
			switch message.Type {
			case Connect:
				//fmt.Println("New player connected from ", message.from.conn.RemoteAddr())
				break
			case JoinRoom:
				m := JoinRoomMessage{}
				if err := json.Unmarshal(message.Content, &m); err != nil {
					message.from.send <- serverErrorHelper("Bad format")
				} else {
					h.joinRoom(message.from, m.Code)
				}
				break
			case CreateRoom:
				h.createRoom(message.from)
				break
			case RoomAction:
				// RoomAction is join/leave room, switch team, send chat message

				// parse the message content as a room message and send to room handler
				if message.from.room != nil {
					rm := RoomActionMessage{}
					rm.from = message.from
					if err := json.Unmarshal(message.Content, &rm); err != nil {
						message.from.send <- serverErrorHelper("Bad RoomActionMessage format")
					} else {
						message.from.room.incomingRoomActions <- rm
					}
				} else {
					message.from.send <- serverErrorHelper("Not in a room")
				}
				break
			case GameAction:
				// related to the trivia gamestate itself

				if message.from.room != nil {
					gam := TriviaGameActionMessage{}
					gam.from = message.from
					if err := json.Unmarshal(message.Content, &gam); err != nil {
						message.from.send <- serverErrorHelper("Bad TriviaGameActionMessage format")

					} else {
						message.from.room.incomingTriviaActions <- gam
					}
				} else {
					message.from.send <- serverErrorHelper("Not in a room")
				}
				break
			default:
				break
			}
		}
	}
}
