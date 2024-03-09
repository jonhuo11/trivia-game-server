package main

import (
	"time"
)

type RoundState int64

const (
	InLimbo RoundState = 0 // time between rounds
	InRound RoundState = 1 // active play time
	InLobby RoundState = 2 // team select
)

// default time per round
const DefaultTriviaRoundTime = 10

// default time between rounds
const DefaultTriviaLimboTime = 5

type TriviaGame struct {
	// state of the current round, limbo or in round
	state RoundState

	// votes, map of player to their selected answer number
	roundVotes map[*Player]int

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

	// assume this is always set
	timer *time.Timer

	// is in test mode?
	debugMode bool

	// round time
	roundTime time.Duration

	// limbo time
	limboTime time.Duration

	// room broadcaster
	roomGameUpdateBroadcaster func(TriviaStateUpdateMessage)
}

func newTriviaGame(broadcaster func(TriviaStateUpdateMessage), debug bool) *TriviaGame {
	return &TriviaGame{
		state:                     InLobby, // team select
		round:                     0,
		timer:                     time.NewTimer(DefaultTriviaLimboTime * time.Second),
		blue:                      make(map[*Player]bool),
		red:                       make(map[*Player]bool),
		blueScore:                 0,
		redScore:                  0,
		debugMode:                 debug,
		roundTime:                 DefaultTriviaRoundTime * time.Second,
		limboTime:                 DefaultTriviaLimboTime * time.Second,
		roomGameUpdateBroadcaster: broadcaster,
	}
}

// reset and start game
func (t *TriviaGame) startGame() {
	t.round = 0
	t.blueScore = 0
	t.redScore = 0
	t.state = InLimbo
	t.goToRoundFromLimbo()
}

/*
Handle incoming player actions and rerouted actions, always runs after the run() cycle
Only 1 action may execute per call
Also handles broadcasting after action completed
*/
func (t *TriviaGame) actionHandlerWithBroadcast(tgam *TriviaGameActionMessage, is *InternalSignal) {
	switch t.state {
	case InLimbo:
		// timer to switch to round
		if is != nil && *is == TriviaGameTimerAlert {
			t.goToRoundFromLimbo()
			t.broadcastGameUpdate(false)
			return
		}
		break
	case InRound:
		// timer to switch to limbo
		if is != nil && *is == TriviaGameTimerAlert {
			t.goToLimboFromRound()
			t.broadcastGameUpdate(false)
		}
		break
	case InLobby:
		// joining teams
		if tgam != nil && tgam.Join != nil {
			if *(tgam.Join) == 0 { // blue
				t.blue[tgam.from] = true
				delete(t.red, tgam.from)
			} else { // red
				t.red[tgam.from] = true
				delete(t.blue, tgam.from)
			}
			t.broadcastGameUpdate(true)
			return
		}

		break
	}
}

// picks a new question from the question bank and sets it as the active question
func (t *TriviaGame) pickNewQuestion(bank interface{}) {
	// TODO
}

// starts a new round
func (t *TriviaGame) goToRoundFromLimbo() {
	if t.state != InLimbo {
		return
	}
	t.round++
	t.state = InRound
	t.timer.Reset(t.roundTime)
}

// enters limbo
func (t *TriviaGame) goToLimboFromRound() {
	if t.state != InRound {
		return
	}
	t.state = InLimbo
	t.timer.Reset(t.limboTime)
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
			blue = append(blue, p.roomname)
		}
		for p := range t.red {
			red = append(red, p.roomname)
		}

		tsum.BlueTeam = &blue
		tsum.RedTeam = &red
	}

	// set state info
	tsum.State = int(t.state)
	tsum.Round = t.round

	t.roomGameUpdateBroadcaster(tsum)
}
