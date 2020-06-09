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
	Score           int  `json:"score"`
	CurrentPosition int  `json:"current_position"`
	Clockwise       bool `json:"clock_wise"`
}

type PlayerBrief struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	HandCardCount int    `json:"hand_card_count"`
	Position      int    `json:"position"`
}

type PlayerDetail struct {
	PlayerBrief
	HandCards []Card `json:"hand_cards"`
}

type server struct {
	mu         *sync.RWMutex
	players    []*player
	maxPlayers int
	games      []*game
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
		games:      make([]*game, 0, maxGames),
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

func (srv *server) findGameById(id string) (*game, error) {
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

	return game.playerJoin(player)
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

func (srv *server) ListPlayersByGame(gameId string) ([]PlayerBrief, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return nil, err
	}

	players := make([]PlayerBrief, 0, len(game.players))
	for _, player := range game.players {
		players = append(players, PlayerBrief{
			Id:            player.id,
			Name:          player.name,
			HandCardCount: len(player.hand),
			Position:      player.position,
		})
	}

	return players, nil
}

func (srv *server) GameInfo(gameId string) (GameDetail, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	game, err := srv.findGameById(gameId)
	if err != nil {
		return GameDetail{}, err
	}
	return GameDetail{
		GameBrief: GameBrief{
			Id:          game.id,
			Name:        game.name,
			State:       game.state,
			PlayerCount: len(game.players),
		},
		Score:           game.score,
		CurrentPosition: game.currentPosition,
		Clockwise:       game.clockwise,
	}, nil
}

func (srv *server) PlayerInfo(playerId string) (PlayerDetail, error) {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	player, err := srv.findPlayerById(playerId)
	if err != nil {
		return PlayerDetail{}, err
	}
	return PlayerDetail{
		PlayerBrief: PlayerBrief{
			Id:            player.id,
			Name:          player.name,
			HandCardCount: len(player.hand),
			Position:      player.position,
		},
		HandCards: player.hand,
	}, nil
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

	return game.playCard(player, cardIndex, cardOption)
}
