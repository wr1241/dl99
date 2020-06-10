package dl99

import (
	"errors"
	"sync"
)

const (
	DefaultMaxPlayers = 600
	DefaultMaxGames   = 100
)

var (
	ErrTooMuchPlayers      = errors.New("too much players")
	ErrTooMuchGames        = errors.New("too much games")
	ErrPlayerNotFound      = errors.New("player not found")
	ErrGameNotFound        = errors.New("game not found")
	ErrYouAreNotInThisGame = errors.New("you are not in this game")
)

type GameBrief struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	State       int    `json:"state"`
	PlayerCount int    `json:"player_count"`
}

type GameDetail struct {
	GameBrief
	Score        int           `json:"score"`
	NextPlayerId string        `json:"next_player_id"`
	Clockwise    bool          `json:"clock_wise"`
	Players      []PlayerBrief `json:"players"`
}

type PlayerBrief struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	HandCardCount int    `json:"hand_card_count"`
}

type PlayerDetail struct {
	PlayerBrief
	HandCards []string `json:"hand_cards"`
}

type server struct {
	mu         *sync.RWMutex
	players    []*player
	maxPlayers int
	games      []*freeBattleGame
	maxGames   int
}

func NewServer(maxPlayers, maxGames int) *server {
	if maxPlayers <= 0 {
		maxPlayers = DefaultMaxPlayers
	}
	if maxGames <= 0 {
		maxGames = DefaultMaxGames
	}
	return &server{
		mu:         &sync.RWMutex{},
		players:    make([]*player, 0, maxPlayers),
		maxPlayers: maxPlayers,
		games:      make([]*freeBattleGame, 0, maxGames),
		maxGames:   maxGames,
	}
}

func (srv *server) findPlayerById(id string) (*player, error) {
	for _, player := range srv.players {
		if player.id == id {
			return player, nil
		}
	}
	return nil, ErrPlayerNotFound
}

func (srv *server) findGameById(id string) (*freeBattleGame, error) {
	for _, game := range srv.games {
		if game.id == id {
			return game, nil
		}
	}
	return nil, ErrGameNotFound
}

func (srv *server) NewPlayer(name string) (string, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if len(srv.players) > srv.maxPlayers {
		return "", ErrTooMuchPlayers
	}

	player := newPlayer(name)
	srv.players = append(srv.players, player)
	return player.id, nil
}

func (srv *server) NewGame(name string) (string, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if len(srv.games) > srv.maxGames {
		return "", ErrTooMuchGames
	}

	game := newGame(name)
	srv.games = append(srv.games, game)
	return game.id, nil
}

func (srv *server) GameBriefs() []GameBrief {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	games := make([]GameBrief, 0, len(srv.games))
	for _, game := range srv.games {
		games = append(games, GameBrief{
			Id:          game.id,
			Name:        game.name,
			State:       game.state,
			PlayerCount: len(game.players),
		})
	}

	return games
}

func (srv *server) JoinGame(gameId string, playerId string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return err
	}

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return err
	}

	return game.join(player)
}

func (srv *server) LeaveGame(gameId string, playerId string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return err
	}

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return err
	}

	return game.leave(player, false)
}

func (srv *server) StartGame(gameId string, playerId string) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return err
	}

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return err
	}

	if player.gameId != game.id {
		return ErrYouAreNotInThisGame
	}

	return game.startGame()
}

func (srv *server) GameInfo(gameId string) (GameDetail, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return GameDetail{}, err
	}

	players := make([]PlayerBrief, 0, len(game.players))
	for _, player := range game.players {
		players = append(players, PlayerBrief{
			Id:            player.id,
			Name:          player.name,
			HandCardCount: len(player.hand),
		})
	}

	return GameDetail{
		GameBrief: GameBrief{
			Id:          game.id,
			Name:        game.name,
			State:       game.state,
			PlayerCount: len(game.players),
		},
		Score:        game.score,
		NextPlayerId: game.nextPlayerId,
		Clockwise:    game.clockwise,
		Players:      players,
	}, nil
}

func (srv *server) PlayerInfo(playerId string) (PlayerDetail, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return PlayerDetail{}, err
	}
	pd := PlayerDetail{
		PlayerBrief: PlayerBrief{
			Id:            player.id,
			Name:          player.name,
			HandCardCount: len(player.hand),
		},
		HandCards: make([]string, 0, len(player.hand)),
	}
	for _, card := range player.hand {
		pd.HandCards = append(pd.HandCards, card.Name())
	}
	return pd, nil
}

func (srv *server) PlayCard(gameId string, playerId string, cardIndex int, cardOption *CardOption) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return err
	}

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return err
	}

	return game.play(player, cardIndex, cardOption)
}

func (srv *server) CleanUpFinishedGame() int {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	toBeDeleteGameIndex := make([]int, 0, len(srv.games))
	for i := len(srv.games) - 1; i >= 0; i-- {
		if srv.games[i].state == GameFinished {
			toBeDeleteGameIndex = append(toBeDeleteGameIndex, i)
		}
	}
	count := 0
	for _, index := range toBeDeleteGameIndex {
		srv.games = append(srv.games[:index], srv.games[index+1:]...)
		count++
	}
	return count
}
