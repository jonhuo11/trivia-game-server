package main

import (
	"testing"
)

func TestTimerEndsRound (t *testing.T) {
	trivia := newTriviaState()
	trivia.goToRoundFromLimbo()
	done := make(chan bool, 1)
	go func() {
		for {
			trivia.run()
			if trivia.roundState == Limbo {
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
	if trivia.roundState != Limbo {
		t.Errorf("State is not Limbo")
	}
}

