package dl99

const (
	defaultPlayerName = "Bravo Player"
)

type player struct {
	id       string
	name     string
	hand     []Card
	gameId   string
	position int
}

func newPlayer(name string) *player {
	if name == "" {
		name = defaultPlayerName
	}
	return &player{
		id:       randomId(),
		name:     name,
		position: -1,
	}
}

func (player player) inGame() bool {
	return player.gameId != ""
}

func (player *player) leaveGame() {
	player.hand = nil
	player.gameId = ""
	player.position = -1
}
