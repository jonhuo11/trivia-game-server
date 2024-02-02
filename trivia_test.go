package main

import (
	"testing"
)

func TestTimerEndsRound(t *testing.T) {
	trivia := newTriviaGame(true)
	trivia.goToRoundFromLimbo()
	done := make(chan bool, 1)
	go func() {
		for {
			trivia.run()
			if trivia.roundState == InLimbo {
				done <- true
				break
			}
		}
	}()
	<-done
	if trivia.round != 1 {
		t.Errorf("Round did not increment from 0 to 1")
	}
	if trivia.timer != 0 {
		t.Errorf("Timer did not reach 0")
	}
	if trivia.roundState != InLimbo {
		t.Errorf("State is not Limbo")
	}
}

func TestTimerStartsRound(t *testing.T) {
	trivia := newTriviaGame(true)
	trivia.endRoundAndGoToLimbo()
	if trivia.roundState != InLimbo {
		t.Fatalf("Expected to be InLimbo")
	}
	for {
		trivia.run()
		if trivia.roundState == InRound {
			break
		}
	}
}
