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

	// room code
	code string

	// mapped to the player number within that room, player 0 is owner
	players map[*Player]int

	// the next player number
	playernum int

	// chat logs
	chat []string

	// state
	state RoomState

	// room + game are closely related, no need for game's own goroutine
	game *TriviaGame

	// client actions for room (chat, join)
	incomingRoomActions chan RoomActionMessage

	// client actions for game (teams, voting, etc)
	incomingTriviaActions chan TriviaGameActionMessage

	// is debugMode
	debugMode bool
}

// room creator helper
func newRoom(debug bool) Room {
	return Room{
		players:               make(map[*Player]int),
		game:                  newTriviaGame(debug),
		incomingRoomActions:   make(chan RoomActionMessage, 1),
		incomingTriviaActions: make(chan TriviaGameActionMessage, 1),
		debugMode: debug,
	}
}

// room loop, main logic here
func (r *Room) run() {
	// game logic
	if r.state == InGame {
		r.game.run() // one game loop
	}

	// read incoming inputs
	select {
	case ram := <-r.incomingRoomActions:
		// chat?
		if ram.Chat != nil {
			r.writeChat(fmt.Sprintf("Player %d: %s", r.players[ram.from], *ram.Chat))
		}

		// start game action handler
		if ram.Start != nil {
			if r.state == InGame {
				ram.from.send <- serverErrorHelper("Game already started")
			} else {
				r.startGame()
			}
		}

		// leave the room
		if ram.Leave != nil && *(ram.Leave) {
			r.removePlayer(ram.from)
		}

		// broadcast updates
		r.broadcastRoomUpdate()
	case tgam := <-r.incomingTriviaActions:
		// route incoming game actions to the trivia handler
		// TODO add error return channel
		r.game.playerAction(tgam)
	case tsum := <-r.game.outgoingTriviaStateUpdateMessages:
		// broadcast to all players the game related update
		r.broadcastGameUpdate(tsum)
	default:
		break
	}
}

// launches trivia game
func (r *Room) startGame() {
	r.game.startGame()
	r.writeChat("Starting new game...")
}

// joins a player to the room
func (r *Room) join(p *Player) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.players[p] = r.playernum
	r.playernum++
	p.room = r
}

// remove player from room and also game team
func (r *Room) removePlayer(player *Player) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, in := r.players[player]; in {
		player.room = nil
		delete(r.players, player)

		delete(r.game.blue, player)
		delete(r.game.red, player)
	}
}

func (r *Room) writeChat(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chat = append(r.chat, msg)
	//fmt.Println(r.chat)
}

// lets clients know about room updates
func (r *Room) broadcastRoomUpdate() {
	if r.debugMode {
		return
	}
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

// lets clients know about game updates
func (r *Room) broadcastGameUpdate(tsum TriviaStateUpdateMessage) {
	if r.debugMode {
		return
	}

	r.mu.Lock() // Lock since r.players is being accessed
	defer r.mu.Unlock()
	for p := range r.players {
		str, _ := json.Marshal(tsum)
		p.send <- OutgoingMessage{
			Type:    TriviaGameUpdate,
			Content: str,
		}
	}
}
