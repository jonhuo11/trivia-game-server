package main

import (
	"encoding/json"
	"fmt"
	"time"
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
	case ram := <-r.incomingRoomActions: // TODO refactor to use action codes
		// chat?
		if ram.Chat != nil {
			r.writeChat(fmt.Sprintf("%s: %s", ram.from.id, *ram.Chat))
		}

		// only the owner can start new games
		if ram.Start != nil {
			v, ok := r.players[ram.from]

			if r.game.state == InRound || r.game.state == InLimbo {
				r.sendErrorTo(ram.from, "Game already started")
			} else if ok && v == 0 {
				r.startGame()
			} else {
				r.sendErrorTo(ram.from, "Only the owner can start a match")
			}
			r.broadcastRoomUpdate(false)
			break
		}

		// join the room, can happen at any time
		if ram.Join != nil && *(ram.Join) {
			r.join(ram.from)
			// notify all players that someone joined
			r.broadcastRoomUpdate(false)
			// notify new player of game state, no need to do room update since broadcastRoomUpdate does this already
			// TODO move this code which gathers all fields for a TSUM struct into a separate helper
			bl, red := r.game.teamIdLists()
			tsum := &TriviaStateUpdateMessage{
				Type:        TSUTSyncNew,
				BlueTeamIds: bl,
				RedTeamIds:  red,
				State:       r.game.state,
				RoundTime: int64(r.game.roundTime / time.Second),
				LimboTime: int64(r.game.limboTime / time.Second),
				StartupTime: int64(r.game.roundTime / time.Second),
			}
			if r.game.activeQuestion != nil {
				tsum.Question = &r.game.activeQuestion.Q
				tsum.Answers = &r.game.activeQuestion.A

				// sync votes, but they are hidden
				votes := []PlayerVote{}
				for p := range r.game.roundVotes {
					_, inblue := r.game.blue[p]
					_, inred := r.game.red[p]
					if inblue || inred {
						votes = append(votes, PlayerVote{p.id, HiddenVoteSelected})
					}
				}
				tsum.Votes = votes

			}
			r.syncPlayer(ram.from, nil, tsum)
			break
		}

		// leave the room, can happen at any time
		if ram.Leave != nil && *(ram.Leave) {
			r.removePlayer(ram.from)
			blue, red := r.game.teamIdLists()
			// notify all players that someone left
			r.broadcastGameUpdate(TriviaStateUpdateMessage{
				Type:        TSUTTeam,
				BlueTeamIds: blue,
				RedTeamIds:  red,
			})
			r.broadcastRoomUpdate(false)
			break
		}
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
	p.id = fmt.Sprintf("Player %d", r.players[p])
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
		playerlist = append(playerlist, p.id)
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
	og, _ := outgoing(RoomUpdate, rum)
	for player := range r.players {
		player.send <- *og
	}
}

// lets clients know about game updates
func (r *Room) broadcastGameUpdate(tsum TriviaStateUpdateMessage) {
	if r.debugMode {
		return
	}

	og, _ := outgoing(TriviaGameUpdate, tsum)
	for p := range r.players {
		p.send <- *og
	}
}

// a new client joined, sync its data
func (r *Room) syncPlayer(to *Player, rum *RoomUpdateMessage, tsum *TriviaStateUpdateMessage) {
	if r.debugMode || to == nil {
		return
	}

	if rum != nil {
		og, _ := outgoing(RoomUpdate, rum)
		to.send <- *og
	}

	if tsum != nil {
		og, _ := outgoing(TriviaGameUpdate, tsum)
		to.send <- *og
	}

}

// helper
func outgoing(smt ServerMessageType, data interface{}) (*OutgoingMessage, error) {
	if str, err := json.Marshal(data); err != nil {
		return nil, err
	} else {
		return &OutgoingMessage{
			Type:    smt,
			Content: str,
		}, nil
	}
}
