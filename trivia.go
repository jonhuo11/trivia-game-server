package main

import (
	"time"
)

type RoundState int64

const (
	Limbo RoundState = 0
	Round RoundState = 1
)

// default time per round
const TriviaRoundTime = 10

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

	// channel for sending updates to be broadcasted
	outgoingTriviaStateUpdateMessages chan TriviaStateUpdateMessage
}

func newTriviaState() *TriviaGame {
	return &TriviaGame{
		roundState: Limbo,
		round:      0,
		timer:      20,
		timerTicker: time.NewTicker(time.Second),
		blue:       make(map[*Player]bool),
		red:        make(map[*Player]bool),
		blueScore:  0,
		redScore:   0,
		question:   "",
		answer:     "",
		outgoingTriviaStateUpdateMessages: make(chan TriviaStateUpdateMessage, 1),
	}
}

// reset and start game
func (t *TriviaGame) startGame() {

}

// one update cycle for the game, send updates to room
func (t *TriviaGame) run() {
	switch t.roundState {
	case Round:
		{
			select {
			case <-t.timerTicker.C:
				//fmt.Println("timer")
				t.timer--
				break
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
	case Limbo:
		{
			// await updates from clients while in limbo

			break
		}
	}
}

// handle incoming player actions
func (t *TriviaGame) playerAction(tgam TriviaGameActionMessage) {
	// joining teams
	if t.roundState == Limbo && tgam.Join != nil {
		// TODO
	}
}

// picks a new question from the question bank and sets it as the active question
func (t *TriviaGame) pickNewQuestion(bank interface{}) {

}

// starts a new round
func (t *TriviaGame) goToRoundFromLimbo() {
	t.round++
	t.timer = TriviaRoundTime
	t.timerTicker.Reset(time.Second)
	t.roundState = Round
}

// ends round
func (t *TriviaGame) endRoundAndGoToLimbo() {
	t.timerTicker.Stop()
	t.roundState = Limbo
}


func (t *TriviaGame) broadcastGameUpdate(updateTeams bool) {
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

	// TODO calculate which updates to broadcast

	t.outgoingTriviaStateUpdateMessages <- tsum
}
