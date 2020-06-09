package dl99

import (
	"errors"
	"log"
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
	ErrYouAreNotCurrentPlayer = errors.New("you are not current player")
	ErrInvalidHandCard        = errors.New("invalid hand card")
)

const (
	deadlineScore        = 99
	minPlayers           = 3
	initialHandCardCount = 5
)

const (
	GameCreated = iota
	GameStarted
	GameFinished
)

const (
	defaultGameName = "Wonderful Game"
)

type game struct {
	mu              *sync.Mutex
	id              string
	name            string
	players         []*player
	score           int
	currentPosition int
	state           int
	rng             *rand.Rand

	deck     []Card
	deadwood []Card

	// if Players order is clockwise, then CurrentPlayerIndex will add by 1
	// if Players order is counterclockwise, then CurrentPlayerIndex will sub by 1
	clockwise bool
}

func newGame(name string) *game {
	if name == "" {
		name = defaultGameName
	}
	game := &game{
		mu:              &sync.Mutex{},
		id:              randomId(),
		name:            name,
		players:         make([]*player, 0, minPlayers),
		score:           0,
		currentPosition: 0,
		state:           GameCreated,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
		clockwise:       true,
	}
	return game
}

func (game *game) playerJoin(player *player) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameCreated {
		return ErrInvalidGameState
	}

	if player.inGame() {
		return ErrPlayerAlreadyJoined
	}

	player.gameId = game.id
	player.position = len(game.players)
	game.players = append(game.players, player)
	log.Printf("player [%s] joined game [%s]\n", player.name, game.name)
	return nil
}

func (game *game) startGame() error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameCreated {
		return ErrInvalidGameState
	}

	if len(game.players) < minPlayers {
		return ErrInSufficientPlayers
	}

	decks := (len(game.players) + 1) / 2
	cards := make([]Card, 0, len(freeBattleDeadline99Deck)*decks)
	for i := 0; i < decks; i++ {
		cards = append(cards, freeBattleDeadline99Deck...)
	}
	shuffle(cards)

	game.deck = cards
	game.deadwood = make([]Card, 0, len(game.deck))
	log.Printf("game [%s] got deck\n", game.name)
	for i := 0; i < len(game.deck); i++ {
		switch i % 10 {
		case 1:
			log.Printf("[")
		case 0:
			log.Println("]")
		}
		log.Printf("%s,", game.deck[i].Name())
	}
	log.Println("\n=====================================")

	for _, player := range game.players {
		if player.hand == nil {
			player.hand = make([]Card, 0, initialHandCardCount)
		} else {
			player.hand = player.hand[:0]
		}
		for i := 0; i < initialHandCardCount; i++ {
			player.hand = append(player.hand, game.drawCard())
		}
	}

	game.state = GameStarted
	log.Printf("game [%s] started\n", game.name)

	return nil
}

func (game *game) playCard(currentPlayer *player, handCardIndex int, cardOption *CardOption) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameStarted {
		return ErrInvalidGameState
	}

	currentPlayerChecked := false
	for _, p := range game.players {
		if p.id == currentPlayer.id && game.currentPosition == currentPlayer.position {
			currentPlayerChecked = true
			break
		}
	}
	if !currentPlayerChecked {
		return ErrYouAreNotCurrentPlayer
	}

	var card Card
	if 0 <= handCardIndex && handCardIndex < len(currentPlayer.hand) {
		card = currentPlayer.hand[handCardIndex]
		currentPlayer.hand = append(currentPlayer.hand[:handCardIndex], currentPlayer.hand[handCardIndex+1:]...)
		game.deadwood = append(game.deadwood, card)
		log.Printf("player [%s] play card [%s] in game [%s]\n", currentPlayer.name, card.Name(), game.name)
	} else {
		return ErrInvalidHandCard
	}

	// Game logic
	tempScore := game.score
	skipDraw := false
	skipNextPosition := false
	switch card.Rank() {
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
		tempScore = deadlineScore
	case RankAce:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for i, p := range game.players {
			if p.id == cardOption.RankAceChangeNextPlayer {
				game.currentPosition = i
				skipNextPosition = true
				break
			}
		}
	case Rank8:
		game.clockwise = !game.clockwise
	case RankJack:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for _, p := range game.players {
			if p.id == cardOption.RankJackDrawOneCardFromPlayer {
				cardIndex := game.rng.Intn(len(p.hand))
				drewCard := p.hand[cardIndex]
				p.hand = append(p.hand[:cardIndex], p.hand[cardIndex+1:]...)
				currentPlayer.hand = append(currentPlayer.hand, drewCard)
				break
			}
		}
		skipDraw = true
	case Rank7:
		if cardOption == nil {
			return ErrInvalidCardOption
		}
		for _, p := range game.players {
			if p.id == cardOption.Rank7ChangeAllHandToPlayer {
				currentPlayer.hand, p.hand = p.hand, currentPlayer.hand
				break
			}
		}
		skipDraw = true
	case Rank3, Rank4, Rank5, Rank6, Rank9:
		tempScore += card.Score()
	default:
		return ErrInvalidRank
	}

	if tempScore < 0 {
		tempScore = 0
	}

	if tempScore > deadlineScore {
		for i, p := range game.players {
			if p.id == currentPlayer.id {
				currentPlayer.leaveGame()
				game.players = append(game.players[:i], game.players[i+1:]...)
				log.Printf("player [%s] left game [%s]\n", currentPlayer.name, game.name)
				log.Printf("now we have %d players remained in game [%s]\n", len(game.players), game.name)
				break
			}
		}
		return ErrLose
	} else {
		game.score = tempScore
		log.Printf("game [%s] score = %d\n", game.name, game.score)
	}

	if len(game.players) == 1 {
		game.state = GameFinished
		currentPlayer.leaveGame()
		game.players = game.players[:0]
		log.Printf("player [%s] won in game [%s]\n", currentPlayer.name, game.name)
		return ErrWin
	}

	if !skipDraw {
		card := game.drawCard()
		currentPlayer.hand = append(currentPlayer.hand, card)
		log.Printf("player [%s] draw card [%s] in game [%s]\n", currentPlayer.name, card.Name(), game.name)
	} else {
		log.Printf("player [%s] skipped draw\n", currentPlayer.name)
	}

	if !skipNextPosition {
		if game.clockwise {
			game.currentPosition = (game.currentPosition + 1) % len(game.players)
		} else {
			game.currentPosition = (game.currentPosition - 1) % len(game.players)
		}
		log.Printf("game [%s] next player will be [%s]\n", game.name, game.players[game.currentPosition].name)
	} else {
		log.Printf("game [%s] will not change next position for now\n", game.name)
	}

	return nil
}

func (game *game) shuffle() {
	if len(game.deck) > 0 {
		game.deadwood = append(game.deadwood, game.deck...)
	}
	shuffle(game.deadwood)
	game.deck = game.deck[:0]
	game.deck = append(game.deck, game.deadwood...)
	game.deadwood = game.deadwood[:0]
}

func (game *game) drawCard() Card {
	if len(game.deck) == 0 {
		game.shuffle()
	}
	card := game.deck[0]
	game.deck = game.deck[1:]
	return card
}
