package main

import (
	"fmt"
	"time"
)

type RoundState int64

const (
	InLimbo RoundState = 0
	InRound RoundState = 1
)

// default time per round
const TriviaRoundTime = 10

// default time between rounds
const TriviaLimboTime = 5

type TriviaGame struct {
	// state of the current round, limbo or in round
	roundState RoundState

	// votes, map of player to their selected answer string
	roundVotes map[*Player]string

	// rounds since game started
	round int

	// playerlist
	blue map[*Player]bool

	// playerlist
	red map[*Player]bool

	// score
	blueScore int

	// score
	redScore int

	// current selected question
	question string

	// answer to current selected question
	answer string

	// time/ticks remaining before current round ends
	timer int

	// assume this is always set
	timerTicker *time.Ticker

	// assume this is always set
	delayTimer *time.Timer

	// channel for sending updates to be broadcasted
	outgoingTriviaStateUpdateMessages chan TriviaStateUpdateMessage

	// is in test mode?
	debugMode bool
}

func newTriviaGame(debug bool) *TriviaGame {
	return &TriviaGame{
		roundState:                        InLimbo,
		round:                             0,
		timer:                             20,
		timerTicker:                       time.NewTicker(time.Second),
		delayTimer:                        time.NewTimer(0),
		blue:                              make(map[*Player]bool),
		red:                               make(map[*Player]bool),
		blueScore:                         0,
		redScore:                          0,
		question:                          "",
		answer:                            "",
		outgoingTriviaStateUpdateMessages: make(chan TriviaStateUpdateMessage, 1),
		debugMode:                         debug,
	}
}

// reset and start game
func (t *TriviaGame) startGame() {

}

// one update cycle for the game, send updates to room
func (t *TriviaGame) run() {
	defer t.broadcastGameUpdate(false)
	switch t.roundState {
	case InRound:
		{
			select {
			case <-t.timerTicker.C:
				if t.debugMode {
					fmt.Println("Trivia timerTicker tick")
				}
				t.timer--
			default:
				break
			}

			if t.timer <= 0 {
				// the round ended
				// calculate winner by aggregating team votes, go to limbo, broadcast
				t.endRoundAndGoToLimbo()
			}

			break
		}
	case InLimbo:
		{
			// wait for go to next round
			select {
			case <-t.delayTimer.C:
				t.goToRoundFromLimbo()
			default:
				break
			}

			break
		}
	}
}

// handle incoming player actions
func (t *TriviaGame) playerAction(tgam TriviaGameActionMessage) {
	switch t.roundState {
	case InLimbo:
		// joining teams
		if tgam.Join != nil {
			if *(tgam.Join) == 0 { // blue
				t.blue[tgam.from] = true
				delete(t.red, tgam.from)
			} else { // red
				t.red[tgam.from] = true
				delete(t.blue, tgam.from)
			}
			t.broadcastGameUpdate(true)
		}
		break
	case InRound:
		break
	}
}

// picks a new question from the question bank and sets it as the active question
func (t *TriviaGame) pickNewQuestion(bank interface{}) {
	// TODO
}

// starts a new round and enters round state immediately
func (t *TriviaGame) goToRoundFromLimbo() {
	if t.roundState != InLimbo {
		return
	}
	t.round++
	t.timer = TriviaRoundTime
	t.timerTicker.Reset(time.Second)
	t.roundState = InRound
}

// ends round and enters limbo state immediately
func (t *TriviaGame) endRoundAndGoToLimbo() {
	if t.roundState != InRound {
		return
	}
	t.timerTicker.Stop()
	t.delayTimer.Reset(time.Second * TriviaLimboTime) // exit Limbo and start new round after X sec
	t.roundState = InLimbo
}

func (t *TriviaGame) broadcastGameUpdate(updateTeams bool) {
	if t.debugMode {
		return
	}

	var tsum = TriviaStateUpdateMessage{}
	if updateTeams {
		blue := []string{}
		red := []string{}
		for p := range t.blue {
			blue = append(blue, p.name)
		}
		for p := range t.red {
			red = append(red, p.name)
		}

		tsum.BlueTeam = &blue
		tsum.RedTeam = &red
	}

	// set state info
	tsum.RoundState = int(t.roundState)
	tsum.Round = t.round
	tsum.RoundTimeLeft = t.timer

	t.outgoingTriviaStateUpdateMessages <- tsum
}
