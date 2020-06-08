package dl99

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrInvalidGameState       = errors.New("invalid game state")
	ErrPlayerAlreadyJoined    = errors.New("player already joined")
	ErrInSufficientPlayers    = errors.New("insufficient players")
	ErrLose                   = errors.New("you lose")
	ErrWin                    = errors.New("you win")
	ErrInvalidRank            = errors.New("invalid rank")
	ErrInvalidCardOption      = errors.New("invalid card option")
	ErrYouAreNotCurrentPlayer = errors.New("you are not current Player")
)

const (
	Deadline   = 99
	MinPlayers = 3
)

var (
	IgnoredCards = []Card{
		{Heart, Rank2}, {Diamond, Rank2}, {Club, Rank2}, {Spade, Rank2},
		{Suit: RedJoker}, {Suit: BlackJoker},
	}
)

const (
	gameCreated = iota
	gameStarted
	gameFinished
)

type Game struct {
	Id                 int64     `json:"id"`
	Name               string    `json:"name"`
	Players            []*Player `json:"players"`
	Score              int       `json:"score"`
	CurrentPlayerIndex int       `json:"current_player_index"`

	mu    sync.Mutex
	state int

	deck     []Card
	deadwood []Card

	// if Players order is clockwise, then CurrentPlayerIndex will add by 1
	// if Players order is counterclockwise, then CurrentPlayerIndex will sub by 1
	clockwise bool
}

func NewGame(id int64, name string) *Game {
	return &Game{
		Id:        id,
		Name:      name,
		state:     gameCreated,
		clockwise: true,
	}
}

func (game *Game) PlayerJoin(player *Player) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != gameCreated {
		return ErrInvalidGameState
	}

	for _, somePlayer := range game.Players {
		if somePlayer.Id == player.Id {
			return ErrPlayerAlreadyJoined
		}
	}
	game.Players = append(game.Players, player)
	return nil
}

func (game *Game) StartGame() error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != gameCreated {
		return ErrInvalidGameState
	}

	if len(game.Players) < MinPlayers {
		return ErrInSufficientPlayers
	}

	setsOfCards := (len(game.Players) + 1) / 2
	cards := make([]Card, 0, StandardNumbersOfPoker*setsOfCards)
	for i := 0; i < setsOfCards; i++ {
		cards = append(cards, ADeckOfCards(IgnoredCards...)...)
	}
	Shuffle(cards)

	game.deck = cards
	game.deadwood = make([]Card, 0, len(game.deck))

	for _, player := range game.Players {
		if player.Hand == nil {
			player.Hand = make([]Card, 0, 5)
		} else {
			player.Hand = player.Hand[:0]
		}
		for i := 0; i < 5; i++ {
			player.Hand = append(player.Hand, game.drawCard())
		}
	}

	game.state = gameStarted

	return nil
}

func (game *Game) PlayCard(player *Player, card Card, cardOption *CardOption) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != gameStarted {
		return ErrInvalidGameState
	}

	if !game.checkCurrentPlayer(player) {
		return ErrYouAreNotCurrentPlayer
	}

	game.deadwood = append(game.deadwood, card)

	// Game logic
	tempScore := game.Score
	skipDraw := false
	switch card.Rank {
	case Rank10:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		if cardOption.Rank10Add {
			tempScore += 10
		} else {
			tempScore -= 10
		}
	case RankQueen:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		if cardOption.RankQueenAdd {
			tempScore += 20
		} else {
			tempScore -= 20
		}
	case RankKing:
		tempScore = Deadline
	case RankAce:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for i, player := range game.Players {
			if player.Id == cardOption.RankAceChangeNextPlayer {
				game.CurrentPlayerIndex = i
				break
			}
		}
	case Rank8:
		game.clockwise = !game.clockwise
	case RankJack:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for _, somePlayer := range game.Players {
			if somePlayer.Id == cardOption.RankJackDrawOneCardFromPlayer {
				cardIndex := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(somePlayer.Hand))
				drewCard := somePlayer.Hand[cardIndex]
				somePlayer.Hand = append(somePlayer.Hand[:cardIndex], somePlayer.Hand[cardIndex+1:]...)
				player.Hand = append(player.Hand, drewCard)
				break
			}
		}
		skipDraw = true
	case Rank7:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for _, somePlayer := range game.Players {
			if somePlayer.Id == cardOption.Rank7ChangeAllHandToPlayer {
				player.Hand, somePlayer.Hand = somePlayer.Hand, player.Hand
				break
			}
		}
		skipDraw = true
	case Rank3:
		tempScore += 3
	case Rank4:
		tempScore += 4
	case Rank5:
		tempScore += 5
	case Rank6:
		tempScore += 6
	case Rank9:
		tempScore += 9
	default:
		return ErrInvalidRank
	}

	if tempScore > Deadline {
		for i, p := range game.Players {
			if p.Id == player.Id {
				game.Players = append(game.Players[:i], game.Players[i+1:]...)
				break
			}
		}
		return ErrLose
	} else {
		game.Score = tempScore
	}

	if len(game.Players) == 1 {
		game.state = gameFinished
		return ErrWin
	}

	if !skipDraw {
		player.Hand = append(player.Hand, game.drawCard())
	}

	if game.clockwise {
		game.CurrentPlayerIndex = (game.CurrentPlayerIndex + 1) % len(game.Players)
	} else {
		game.CurrentPlayerIndex = (game.CurrentPlayerIndex - 1) % len(game.Players)
	}

	return nil
}

func (game *Game) checkCurrentPlayer(player *Player) bool {
	for i, p := range game.Players {
		if p.Id == player.Id {
			return game.CurrentPlayerIndex == i
		}
	}
	return false
}

func (game *Game) shuffle() {
	game.deadwood = append(game.deadwood, game.deck...)
	Shuffle(game.deadwood)
	game.deck = game.deck[:0]
	game.deck = append(game.deck, game.deadwood...)
	game.deadwood = game.deadwood[:0]
}

func (game *Game) drawCard() Card {
	if len(game.deck) == 0 {
		game.shuffle()
	}

	card := game.deck[0]
	game.deck = game.deck[1:]
	return card
}
