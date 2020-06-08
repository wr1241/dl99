package dl99

import "errors"

type Player struct {
	Id   int64  `json:"Id"`
	Name string `json:"name"`
	Hand []Card `json:"hand"`
}

func NewPlayer(id int64, name string) *Player {
	return &Player{
		Id:   id,
		Name: name,
	}
}

func (player *Player) Play(game *Game, cardIndex int, cardOption *CardOption) error {
	if cardIndex >= len(player.Hand) {
		return errors.New("invalid card index")
	}

	card := player.Hand[cardIndex]
	player.Hand = append(player.Hand[:cardIndex], player.Hand[cardIndex+1:]...)
	return game.PlayCard(player, card, cardOption)
}
