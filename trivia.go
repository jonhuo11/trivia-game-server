package main

import (
	"math/rand"
	"time"
)

type RoundState int64

const (
	InLimbo RoundState = 0 // time between rounds
	InRound RoundState = 1 // active play time
	InLobby RoundState = 2 // team select
)

// default time per round (seconds)
const DefaultTriviaRoundTime = 20

// default time between rounds
const DefaultTriviaLimboTime = 10

// default time to transition from InLobby to InRound
const DefaultTriviaStartupTime = 5

type roomBroadcaster func(TriviaStateUpdateMessage)

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

	// bank of questions
	bank TriviaBank

	// current question
	activeQuestion *TriviaQuestion

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

	// startup time
	startupTime time.Duration

	// room broadcaster
	roomGameUpdateBroadcaster roomBroadcaster
}

func newTriviaGame(broadcaster roomBroadcaster, debug bool) *TriviaGame {
	bank, _ := loadTriviaBankDefault()
	timer := time.NewTimer(0)
	timer.Stop()
	return &TriviaGame{
		state:                     InLobby, // team select
		round:                     0,
		timer:                     timer,
		blue:                      make(map[*Player]bool),
		red:                       make(map[*Player]bool),
		bank:                      bank,
		activeQuestion:            nil,
		blueScore:                 0,
		redScore:                  0,
		debugMode:                 debug,
		roundTime:                 DefaultTriviaRoundTime * time.Second,
		limboTime:                 DefaultTriviaLimboTime * time.Second,
		startupTime:               DefaultTriviaStartupTime * time.Second,
		roomGameUpdateBroadcaster: broadcaster,
	}
}

// reset and start game
func (t *TriviaGame) startGame() {
	t.round = 0
	t.blueScore = 0
	t.redScore = 0

	t.broadcastGameUpdate(TriviaStateUpdateMessage{
		Type:        TSUTStartup,
		RoundTime:   int64(t.roundTime / time.Second),
		LimboTime:   int64(t.roundTime / time.Second),
		StartupTime: int64(t.startupTime / time.Second),
	})

	// startup timer
	t.timer.Reset(t.startupTime)
}

/*
Handle incoming player actions and rerouted actions from room.go (uses InternalSignal)
Always runs after the run() cycle
Only 1 action may execute per call
Also handles broadcasting after action completed
*/
func (t *TriviaGame) actionHandlerWithBroadcast(tgam *TriviaGameActionMessage, is *InternalSignal) {
	switch t.state {
	case InLimbo:
		// timer to switch to round
		if is != nil && *is == TriviaGameTimerAlert {
			t.goToRoundFromLimboWithBroadcast()
			return
		}
		break
	case InRound:
		// timer to switch to limbo
		if is != nil && *is == TriviaGameTimerAlert {
			t.goToLimboFromRoundWithBroadcast()
			return
		}
		if tgam != nil && tgam.from != nil {
			// player guess
			if tgam.Type == TGATGuess && tgam.Guess >= 0 && tgam.Guess < len(t.activeQuestion.A) {
				t.roundVotes[tgam.from] = tgam.Guess
			}
		}
		break
	case InLobby:
		// startup timer is done
		if is != nil && *is == TriviaGameTimerAlert {
			t.state = InLimbo
			//fmt.Println("timer")
			t.goToRoundFromLimboWithBroadcast()
			return
		}

		// joining teams
		if tgam != nil && tgam.Type == TGATJoin {
			//fmt.Println("joining teams")
			t.joinTeamWithBroadcast(tgam.Join, tgam.from)
			return
		}
		break
	}
}

// picks a new question from the question bank and sets it as the active question
func (t *TriviaGame) pickNewQuestion() {
	q := t.bank.Questions[rand.Intn(len(t.bank.Questions))]
	t.activeQuestion = &q
}

// starts a new round
func (t *TriviaGame) goToRoundFromLimboWithBroadcast() {
	if t.state != InLimbo {
		return
	}
	t.round++
	t.state = InRound
	t.timer.Reset(t.roundTime)
	t.pickNewQuestion()
	t.broadcastGameUpdate(TriviaStateUpdateMessage{
		Type:     TSUTGoToRoundFromLimbo,
		Round:    t.round,
		State:    t.state,
		Question: &t.activeQuestion.Q,
		Answers:  &t.activeQuestion.A,
	})
}

// enters limbo
func (t *TriviaGame) goToLimboFromRoundWithBroadcast() {
	if t.state != InRound {
		return
	}
	t.state = InLimbo
	t.timer.Reset(t.limboTime)
	t.broadcastGameUpdate(TriviaStateUpdateMessage{
		Type:  TSUTGoToLimboFromRound,
		State: t.state,
	})
}

func (t *TriviaGame) teamIdLists() (blue []string, red []string) {
	b := []string{}
	r := []string{}
	for k := range t.blue {
		b = append(b, k.id)
	}
	for k := range t.red {
		r = append(r, k.id)
	}
	return b, r
}

func (t *TriviaGame) joinTeamWithBroadcast(team int, who *Player) {
	if team == 0 { // blue
		t.blue[who] = true
		delete(t.red, who)
	} else { // red
		t.red[who] = true
		delete(t.blue, who)
	}
	b, r := t.teamIdLists()
	t.broadcastGameUpdate(TriviaStateUpdateMessage{
		Type:        TSUTTeam,
		BlueTeamIds: b,
		RedTeamIds:  r,
	})
}

func (t *TriviaGame) applyPlayerGuessWithBroadcast(guess int, who *Player) {

}

func (t *TriviaGame) broadcastGameUpdate(tsum TriviaStateUpdateMessage) {
	if t.debugMode {
		return
	}

	t.roomGameUpdateBroadcaster(tsum)
}
