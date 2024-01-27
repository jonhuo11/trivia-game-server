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
	round     int
	timer     int
	blue      map[*Player]bool
	red       map[*Player]bool
	blueScore int
	redScore  int
	question  string
	answer    string
}

func newTriviaState () TriviaState {
	return TriviaState {
		round: 0,
		timer: 20,
		blue: make(map[*Player]bool),
		red: make(map[*Player]bool),
		blueScore: 0,
		redScore: 0,
		question: "",
		answer: "",
	}
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

			// broadcast updates
			r.broadcastRoomUpdate()
			break
		case tgam := <-r.incomingTriviaActions:
			fmt.Println("Got TriviaGameAction from a player")
			switch(r.state) {
			case Lobby:
				// team select
				if tgam.Join != nil {
					r.joinTeam(tgam.from, *tgam.Join)
				}
				r.broadcastGameUpdate()
				break
			case InGame:
				break
			default:
				break
			}
			break
		}
	}
}

// 0 is blue 1 is red
func (r *Room) joinTeam(player* Player, team int) {
	r.mu.Lock()
	if team == 0 {
		r.gamestate.blue[player] = true
		delete(r.gamestate.red, player)
	} else {
		r.gamestate.red[player] = true
		delete(r.gamestate.blue, player)
	}
	r.mu.Unlock()
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

func (r *Room) broadcastGameUpdate() {
	if r == nil {return}
	r.mu.Lock()
	defer r.mu.Unlock()

	blue := []string{}
	red := []string{}
	for p := range r.gamestate.blue {
		blue = append(blue, p.name)
	}
	for p := range r.gamestate.red {
		red = append(red, p.name)
	}

	tsum := TriviaStateUpdateMessage{
		BlueTeam: blue,
		RedTeam: red,
	}
	str, _ := json.Marshal(tsum)
	for player := range r.players{
		player.send <- OutgoingMessage{
			Type: TriviaGameUpdate,
			Content: str,
		}
	}
}
