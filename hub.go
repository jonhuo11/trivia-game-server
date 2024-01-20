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

// lets clients know about room updates
func (h *Hub) broadcastRoomUpdate(r *Room) {
	if r == nil {
		return
	}
	newplayers := []string{}
	for p, _ := range r.Players {
		newplayers = append(newplayers, p.name)
	}
	rum := RoomUpdateMessage{
		Code:    r.Code,
		Players: newplayers,
	}
	str, _ := json.Marshal(rum)
	for player, _ := range r.Players {
		player.send <- OutgoingMessage{
			Type:    RoomUpdate,
			Content: str,
		}
	}
}

func (h *Hub) joinRoom(p *Player, code string) *Room {
	if p.room != nil {
		p.send <- ServerErrorHelper("this player is already in a room")
		return nil
	}
	if room, in := h.rooms[code]; !in {
		p.send <- ServerErrorHelper("this room does not exist")
		return nil
	} else {
		room.RegisterPlayer(p)
		p.room = room
		return room
	}
}

func (h *Hub) createRoom(creator *Player) *Room {
	if creator.room != nil {
		creator.send <- ServerErrorHelper("this player is already in a room")
		return nil
	}
	id := uuid.New().String()
	h.rooms[id] = &Room{
		Code:    id,
		Players: map[*Player]bool{creator: true},
	}
	creator.room = h.rooms[id]
	fmt.Println(h.rooms)
	return h.rooms[id]
}

func (h *Hub) run() {
	for {
		select {
		case player := <-h.register:
			h.players[player] = true
		case player := <-h.unregister:
			delete(h.players, player)
			//fmt.Println(h.rooms)
			if player.room != nil {
				if _, in := h.rooms[player.room.Code]; in {
					delete(h.rooms[player.room.Code].Players, player)
					//fmt.Println(h.rooms[player.room.Code])
					h.broadcastRoomUpdate(h.rooms[player.room.Code])
				}
			}
			close(player.send)
			fmt.Println("Unregistered client and removed from room")
			break
		case message := <-h.incoming:
			switch message.Type {
			case Connect:
				fmt.Println(message.From.conn.RemoteAddr(), "connected")
				break
			case JoinRoom:
				fmt.Println("Try to join room")
				m := JoinRoomMessage{}
				if err := json.Unmarshal(message.Content, &m); err != nil {
					message.From.send <- ServerErrorHelper("Bad format")
				} else {
					h.broadcastRoomUpdate(h.joinRoom(message.From, m.Code))
				}
				break
			case CreateRoom:
				fmt.Println("Try to create room")
				h.broadcastRoomUpdate(h.createRoom(message.From))
				break
			default:
				break
			}
		}
	}
}
