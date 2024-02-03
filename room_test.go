package main

import (
	"fmt"
	"testing"
)

func TestStart(t *testing.T) {
	room := newRoom("test", true)
	if room.game.state == InRound {
		t.Fatalf("Room should start in Limbo")
	}
	pl0 := &Player{} // admin
	pl1 := &Player{} // not admin
	room.join(pl0)
	room.join(pl1)

	// only admin can start room
	sm := RoomActionMessage{}
	sm.from = pl1
	sm.Start = boolPtr(true)
	room.incomingRoomActions <- sm
	room.run()
	if room.game.state == InRound {
		t.Fatalf("Game was started but only admin should be able to start game")
	}

	// admin starts room successfully
	sm.from = pl0
	room.incomingRoomActions <- sm
	room.run()
	if room.game.state == InLimbo {
		t.Fatalf("Admin should be able to start game")
	}
}

func TestLeave(t *testing.T) {
	room := newRoom("test", true)
	room.startGame()

	// join player to the room
	pl := Player{}
	room.join(&pl)

	// send leave room message from player
	ram := RoomActionMessage{}
	ram.from = &pl
	tmp := true
	ram.Leave = &tmp
	room.incomingRoomActions <- ram

	room.run()
	//fmt.Println(room.players)

	if len(room.players) != 0 || len(room.game.blue) != 0 || len(room.game.red) != 0 {
		t.Errorf("Player list should be empty")
	}

}

func TestRoundsRotateFromRoundToLimbo(t *testing.T) {
	room := newRoom("test", true)
	pl := &Player{}
	room.join(pl)
	room.startGame()
	if room.game.state != InRound {
		t.Fatalf("Should be in Round before starting flip flop test")
	}
	fmt.Println("Waiting for round timer...")
	room.run() //  go until round timer, should switch to limbo
	if room.game.state != InLimbo {
		t.Fatalf("Did not go to Limbo after timer went off")
	}
	fmt.Println("Waiting for limbo timer...")
	room.run() //  go until round timer, should switch to round
	if room.game.state != InRound {
		t.Fatalf("Did not go to Round after timer went off")
	}
	fmt.Println("Finished 1 round to limbo rotation")

}
