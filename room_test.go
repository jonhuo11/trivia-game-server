package main

import (
	"testing"
)

func boolPtr(v bool) *bool {
	x := v
	return &x
}

func TestStart(t *testing.T) {
	room := newRoom(true)
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
	room := newRoom(true)
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

