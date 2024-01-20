package main

type Room struct {
	Code string

	Players map[*Player]bool
}

func (r *Room) RegisterPlayer(player *Player) {
	r.Players[player] = true
}
