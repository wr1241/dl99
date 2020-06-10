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
	ErrPlayerNotInThisGame    = errors.New("player not in this game")
	ErrInSufficientPlayers    = errors.New("insufficient players")
	ErrInsufficientCards      = errors.New("insufficient cards")
	ErrLose                   = errors.New("you lose")
	ErrWin                    = errors.New("you win")
	ErrInvalidRank            = errors.New("invalid rank")
	ErrInvalidCardOption      = errors.New("invalid card option")
	ErrYouAreNotCurrentPlayer = errors.New("you are not current player")
	ErrInvalidHandCard        = errors.New("invalid hand card")
)

const (
	deadlineScore        = 99
	minPlayers           = 2
	initialHandCardCount = 5
	defaultGameName      = "Wonderful Game"
	gamePrefix           = "g-"
)

const (
	GameCreated = iota
	GameStarted
	GameFinished
)

type freeBattleGame struct {
	mu           *sync.Mutex
	id           string
	name         string
	players      []*player
	score        int
	nextPlayerId string
	state        int
	rng          *rand.Rand

	deck     []Card
	deadwood []Card

	// if Players order is clockwise, then CurrentPlayerIndex will add by 1
	// if Players order is counterclockwise, then CurrentPlayerIndex will sub by 1
	clockwise bool
}

func newGame(name string) *freeBattleGame {
	if name == "" {
		name = defaultGameName
	}
	game := &freeBattleGame{
		mu:        &sync.Mutex{},
		id:        randomId(gamePrefix),
		name:      name,
		players:   make([]*player, 0, minPlayers),
		state:     GameCreated,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
		clockwise: true,
	}
	return game
}

func (game *freeBattleGame) join(player *player) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameCreated {
		return ErrInvalidGameState
	}

	if player.inGame() {
		return ErrPlayerAlreadyJoined
	}

	player.gameId = game.id
	player.hand = nil
	game.players = append(game.players, player)
	log.Printf("game [%s] player [%s] joined", game.name, player.name)

	return nil
}

func (game *freeBattleGame) leave(player *player, inGame bool) error {
	// 当 inGame 为真时，leave() 方法是在 play() 内部调用的。
	// 已经持有 game 对象内部的锁了，就不需要再持有了。
	// 当 inGame 为假时，leave() 方法是在 API 请求里调用的。
	// 这种情况就应该加锁，以免破坏游戏数据。
	if !inGame {
		game.mu.Lock()
		defer game.mu.Unlock()
	}

	if player.gameId != game.id {
		return ErrPlayerNotInThisGame
	}

	playerCount := len(game.players)
	for i := 0; i < playerCount; i++ {
		if game.players[i].id == player.id {
			if game.state == GameStarted {
				// 如果当前离开的玩家就是下一位该出牌的玩家
				// 就要再计算一次下一位出牌玩家
				// 但是有一个特例，如果只剩最后一位玩家了，就不计算了
				if game.nextPlayerId == player.id && playerCount > 1 {
					if game.clockwise {
						game.nextPlayerId = game.players[(i+1)%playerCount].id
					} else {
						game.nextPlayerId = game.players[(i+playerCount-1)%playerCount].id
					}
				}

				// 离开的玩家的手牌，都要丢进弃牌堆
				game.deadwood = append(game.deadwood, player.hand...)
				player.hand = nil

				// 最后离开的玩家获胜了
				if len(game.players) < 1 {
					game.state = GameFinished
					return ErrWin
				}
			}

			game.players = append(game.players[:i], game.players[i+1:]...)
			player.gameId = ""
			log.Printf("game [%s] player [%s] left, now we have %d players remained",
				game.name, player.name, len(game.players))

			if game.state == GameStarted {
				return ErrLose
			}

			return nil
		}
	}

	panic("player's game id exists, but player not in game's player list")
}

func (game *freeBattleGame) startGame() error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameCreated {
		return ErrInvalidGameState
	}

	playerCount := len(game.players)
	if playerCount < minPlayers {
		return ErrInSufficientPlayers
	}

	setsOfCards := (playerCount + 1) / 2
	game.deck = make([]Card, 0, len(freeBattleDeadline99Deck)*setsOfCards)
	for i := 0; i < setsOfCards; i++ {
		game.deck = append(game.deck, freeBattleDeadline99Deck...)
	}
	shuffle(game.deck)

	game.deadwood = make([]Card, 0, len(game.deck))

	for _, player := range game.players {
		if err := game.drawCard(player, initialHandCardCount); err != nil {
			return err
		}
	}

	game.nextPlayerId = game.players[0].id
	game.state = GameStarted
	log.Printf("game [%s] started", game.name)

	return nil
}

func (game *freeBattleGame) play(currentPlayer *player, handCardIndex int, cardOption *CardOption) error {
	game.mu.Lock()
	defer game.mu.Unlock()

	if game.state != GameStarted {
		return ErrInvalidGameState
	}

	if game.nextPlayerId != currentPlayer.id {
		return ErrYouAreNotCurrentPlayer
	}

	// 当前玩家出牌前，就被别人把手牌取光了
	if len(currentPlayer.hand) == 0 {
		if err := game.leave(currentPlayer, true); err != nil {
			log.Printf("game [%s] player [%s] leave failed: %v", game.name, currentPlayer.name, err)
			return err
		}
		log.Printf("game [%s] player [%s] lose due to have no hand cards", game.name, currentPlayer.name)
		return ErrLose
	}

	var card Card
	if 0 <= handCardIndex && handCardIndex < len(currentPlayer.hand) {
		card = currentPlayer.hand[handCardIndex]
		currentPlayer.hand = append(currentPlayer.hand[:handCardIndex], currentPlayer.hand[handCardIndex+1:]...)
		game.deadwood = append(game.deadwood, card)
		log.Printf("game [%s] player [%s] play card [%s]", game.name, currentPlayer.name, card.Name())
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
		for _, p := range game.players {
			if p.id == cardOption.RankAceChangeNextPlayer {
				game.nextPlayerId = cardOption.RankAceChangeNextPlayer
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
		if err := game.leave(currentPlayer, true); err != nil {
			log.Printf("game [%s] player [%s] leave failed: %v", game.name, currentPlayer.name, err)
			return err
		}
		log.Printf("game [%s] player [%s] lose due to beyond the deadline", game.name, currentPlayer.name)
		return ErrLose
	} else {
		game.score = tempScore
		log.Printf("game [%s] score is %d", game.name, game.score)
	}

	if len(game.players) == 1 {
		if err := game.leave(currentPlayer, true); err != nil {
			log.Printf("game [%s] player [%s] leave failed: %v", game.name, currentPlayer.name, err)
			return err
		}
		game.state = GameFinished
		log.Printf("player [%s] won in game [%s]", currentPlayer.name, game.name)
		return ErrWin
	}

	if !skipDraw {
		if err := game.drawCard(currentPlayer, 1); err != nil {
			log.Printf("game [%s] player [%s] draw card failed: %v", game.name, currentPlayer.name, err)
		}
	} else {
		log.Printf("game [%s] player [%s] skipped draw", game.name, currentPlayer.name)
	}

	if !skipNextPosition {
		playerCount := len(game.players)
		for i := 0; i < len(game.players); i++ {
			if game.players[i].id == currentPlayer.id {
				if game.clockwise {
					game.nextPlayerId = game.players[(i+1)%playerCount].id
				} else {
					game.nextPlayerId = game.players[(i+playerCount-1)%playerCount].id
				}
				break
			}
		}
		log.Printf("game [%s] next player will be [%s]", game.name, game.nextPlayerId)
	} else {
		log.Printf("game [%s] will not change next player for now", game.name)
	}

	return nil
}

func (game *freeBattleGame) recycle() {
	if len(game.deck) > 0 {
		game.deadwood = append(game.deadwood, game.deck...)
		game.deck = game.deck[:0]
	}
	shuffle(game.deadwood)
	game.deck, game.deadwood = game.deadwood, game.deck
}

func (game *freeBattleGame) drawCard(player *player, count int) error {
	if len(game.deck) < count {
		game.recycle()
	}

	if len(game.deck) < count {
		return ErrInsufficientCards
	}

	if player.hand == nil {
		player.hand = make([]Card, 0, count)
	}

	for i := 0; i < count; i++ {
		player.hand = append(player.hand, game.deck[0])
		game.deck = game.deck[1:]
	}
	log.Printf("game [%s] player [%s] drew %d cards, we have %d card in deck",
		game.name, player.name, count, len(game.deck))

	return nil
}
