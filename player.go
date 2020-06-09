package dl99

const (
	defaultPlayerName = "Bravo Player"
	playerIdPrefix    = "p-"
)

type player struct {
	id     string
	name   string
	gameId string
	hand   []Card
}

func newPlayer(name string) *player {
	if name == "" {
		name = defaultPlayerName
	}
	return &player{
		id:         randomId(playerIdPrefix),
		name:       name,
		gameId:     "",
		hand:       nil,
	}
}

func (player player) inGame() bool {
	return player.gameId != ""
}
