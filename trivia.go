package main

import "time"

type RoundState int64

const (
	Limbo RoundState = 0
	Round RoundState = 1
)

const TriviaRoundTime = 10

type TriviaState struct {
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

	// TODO channel for sending
}

func newTriviaState() TriviaState {
	return TriviaState{
		roundState: Limbo,
		round:      0,
		timer:      20,
		blue:       make(map[*Player]bool),
		red:        make(map[*Player]bool),
		blueScore:  0,
		redScore:   0,
		question:   "",
		answer:     "",
	}
}

// one update cycle for the game, send updates to room
func (t *TriviaState) run() {
	switch t.roundState {
	case Round:
		{
			select {
			case <-t.timerTicker.C:
				t.timer--
				break
			default:
				break
			}
			if t.timer <= 0 {
				// the round ended
				// calculate winner by aggregating team votes, go to limbo, broadcast
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

// picks a new question from the question bank and sets it as the active question
func (t *TriviaState) pickNewQuestion(bank interface{}) {

}

// starts a new round
func (t *TriviaState) goToRoundFromLimbo() {

}

// ends round
func (t *TriviaState) endRoundAndGoToLimbo() {

}
