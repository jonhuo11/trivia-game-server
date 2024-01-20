package main

import (
	"encoding/json"
	"fmt"
	"sync"
)

type RoomState int64

const (
	Lobby  RoomState = 0
	InGame RoomState = 1
)

type TriviaState struct {
	Round     int
	Timer     int
	Blue      map[*Player]bool
	Red       map[*Player]bool
	BlueScore int
	RedScore  int
	Question  string
	Answer    string
}

type Room struct {
	mu sync.Mutex

	code string

	// mapped to the player number within that room
	players map[*Player]int

	chat []string

	state RoomState

	gamestate TriviaState

	incomingRoomActions chan RoomActionMessage

	incomingTriviaActions chan TriviaGameActionMessage
}

// room loop, main game logic here
func (r *Room) run() {
	for {
		select {
		case ram := <-r.incomingRoomActions:
			//fmt.Println("Room action received")
			// chat?
			if ram.Chat != nil {
				r.writeChat(fmt.Sprintf("Player %d: %s", r.players[ram.from], *ram.Chat))
			}

			// broadcast an update
			r.broadcastRoomUpdate()
			break
		case tgam := <-r.incomingTriviaActions:
			fmt.Println("Not implemented trivia actions", tgam)
			break
		}
	}
}


func (r *Room) removePlayer(player *Player) {
	r.mu.Lock()
	if _, in := r.players[player]; in {
		player.room = nil
		delete(r.players, player)
	}
	r.mu.Unlock()
}

func (r *Room) writeChat(msg string) {
	r.mu.Lock()
	r.chat = append(r.chat, msg)
	//fmt.Println(r.chat)
	r.mu.Unlock()
}

// lets clients know about room updates
func (r *Room) broadcastRoomUpdate() {
	if r == nil {
		return
	}
	r.mu.Lock()
	playerlist := []string{}
	for p := range r.players {
		playerlist = append(playerlist, p.name)
	}
	rum := RoomUpdateMessage{
		Code:    r.code,
		Players: playerlist,
		Chat:    r.chat,
	}
	str, _ := json.Marshal(rum)
	for player := range r.players {
		player.send <- OutgoingMessage{
			Type:    RoomUpdate,
			Content: str,
		}
	}
	r.mu.Unlock()
}
