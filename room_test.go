package main

import (
	"testing"
)

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
