package main

import (
	"encoding/json"
	"fmt"
)

type Room struct {
	// room code
	code string

	// mapped to the player number within that room, player 0 is owner
	players map[*Player]int

	// the next player number
	playernum int

	// chat logs
	chat []string

	// room + game are closely related, no need for game's own goroutine. Room will control game
	game *TriviaGame

	// client actions for room (chat, join)
	incomingRoomActions chan RoomActionMessage

	// client actions for game (teams, voting, etc)
	incomingTriviaActions chan TriviaGameActionMessage

	// is debugMode
	debugMode bool
}

// room creator helper
func newRoom(id string, debug bool) *Room {
	r := Room{
		players:               make(map[*Player]int),
		incomingRoomActions:   make(chan RoomActionMessage, 1),
		incomingTriviaActions: make(chan TriviaGameActionMessage, 1),
		debugMode:             debug,
		code:                  id,
		chat:                  []string{},
	}
	g := newTriviaGame(r.broadcastGameUpdate, debug)
	r.game = g
	return &r
}

// room loop, main logic here
func (r *Room) run() {
	/*
		Blocks calculations until a signal is detected. Signals are:
		1. Incoming room/game action
		2. Outgoing game update
		3. Round timer
	*/
	select {
	case ram := <-r.incomingRoomActions:
		// chat?
		if ram.Chat != nil {
			r.writeChat(fmt.Sprintf("%s: %s", ram.from.roomname, *ram.Chat))
		}

		// only the owner can start new games
		if ram.Start != nil {
			v, ok := r.players[ram.from]

			if r.game.state == InRound {
				r.sendErrorTo(ram.from, "Game already started")
			} else if ok && v == 0 {
				r.startGame()
			} else {
				r.sendErrorTo(ram.from, "Only the owner can start a match")
			}
		}

		// join the room
		if ram.Join != nil && *(ram.Join) {
			r.join(ram.from)
		}

		// leave the room
		if ram.Leave != nil && *(ram.Leave) {
			r.removePlayer(ram.from)
		}

		// broadcast updates
		r.broadcastRoomUpdate(false)
	case tgam := <-r.incomingTriviaActions:
		// route incoming game actions to the trivia handler
		// TODO add error return channel
		r.game.actionHandlerWithBroadcast(&tgam, nil)
	case <-r.game.timer.C:
		// timer went off, reroute back to game handler
		signal := TriviaGameTimerAlert
		r.game.actionHandlerWithBroadcast(nil, &signal)
	}
}

// send error to player
func (r *Room) sendErrorTo(p *Player, msg string) {
	if r.debugMode {
		return
	}
	p.send <- serverErrorHelper(msg)
}

// launches trivia game
func (r *Room) startGame() {
	r.game.startGame()
	r.writeChat("Starting new game...")
}

// joins a player to the room
func (r *Room) join(p *Player) {
	r.players[p] = r.playernum
	r.playernum++
	p.room = r
	p.roomname = fmt.Sprintf("Player %d", r.players[p])
}

// remove player from room and also game team
func (r *Room) removePlayer(player *Player) {
	if _, in := r.players[player]; in {
		player.room = nil
		delete(r.players, player)

		delete(r.game.blue, player)
		delete(r.game.red, player)
	}
}

func (r *Room) writeChat(msg string) {
	r.chat = append(r.chat, msg)
	//fmt.Println(r.chat)
}

// lets clients know about room updates
func (r *Room) broadcastRoomUpdate(created bool) {
	if r.debugMode {
		return
	}

	playerlist := []string{}
	for p := range r.players {
		playerlist = append(playerlist, p.roomname)
	}
	rum := RoomUpdateMessage{
		Code:    r.code,
		Players: playerlist,
		Chat:    r.chat,
	}
	if created {
		tmp := true
		rum.Created = &tmp
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

	for p := range r.players {
		str, _ := json.Marshal(tsum)
		p.send <- OutgoingMessage{
			Type:    TriviaGameUpdate,
			Content: str,
		}
	}
}
