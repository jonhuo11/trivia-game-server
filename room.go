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

type Room struct {
	mu sync.Mutex

	code string

	// mapped to the player number within that room, player 0 is owner
	players map[*Player]int

	// the next player number
	playernum int

	chat []string

	state RoomState

	// room + game are closely related, no need for game's own goroutine
	game *TriviaGame

	incomingRoomActions chan RoomActionMessage

	incomingTriviaActions chan TriviaGameActionMessage
}

// room loop, main logic here
func (r *Room) run() {
	for {
		select {
		case ram := <-r.incomingRoomActions:
			// chat?
			if ram.Chat != nil {
				r.writeChat(fmt.Sprintf("Player %d: %s", r.players[ram.from], *ram.Chat))
			}

			// broadcast updates
			r.broadcastRoomUpdate()
		case tgam := <-r.incomingTriviaActions:
			// route incoming game actions to the trivia handler
			r.game.playerAction(tgam)
		case tsum := <-r.game.outgoingTriviaStateUpdateMessages:
			// broadcast to all players the game related update
			for p := range r.players {
				str, _ := json.Marshal(tsum)
				p.send <- OutgoingMessage{
					Type: TriviaGameUpdate,
					Content: str,
				}
			}
		default:
			// game logic
		}
	}
}


// launches trivia game
func (r *Room) startGame () {
	
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
	r.mu.Lock()
	defer r.mu.Unlock()
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
}

