package dl99

import (
	"fmt"
	"github.com/google/uuid"
)

type Suit uint8

const (
	_ Suit = iota
	Heart
	Diamond
	Club
	Spade
	RedJoker
	BlackJoker
)

func (suit Suit) Name() string {
	switch suit {
	case Heart:
		return "Heart"
	case Diamond:
		return "Diamond"
	case Club:
		return "Club"
	case Spade:
		return "Spade"
	case RedJoker:
		return "Red Joker"
	case BlackJoker:
		return "Black Joker"
	}
	panic("invalid Suit Name")
}

type Rank uint8

const (
	_ Rank = iota
	RankAce
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
	Rank9
	Rank10
	RankJack
	RankQueen
	RankKing
)

func (rank Rank) Name() string {
	switch rank {
	case RankAce:
		return "Ace"
	case Rank2:
		return "Two"
	case Rank3:
		return "Three"
	case Rank4:
		return "Four"
	case Rank5:
		return "Five"
	case Rank6:
		return "Six"
	case Rank7:
		return "Seven"
	case Rank8:
		return "Eight"
	case Rank9:
		return "Nine"
	case Rank10:
		return "Ten"
	case RankJack:
		return "Jack"
	case RankQueen:
		return "Queen"
	case RankKing:
		return "King"
	}
	panic("invalid Rank Name")
}

type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

func (card Card) Name() string {
	switch card.Suit {
	case RedJoker, BlackJoker:
		return card.Suit.Name()
	default:
		return fmt.Sprintf("%s of %s", card.Rank.Name(), card.Suit.Name())
	}
}

func (card Card) Score() int {
	switch card.Rank {
	case Rank3:
		return 3
	case Rank4:
		return 4
	case Rank5:
		return 5
	case Rank6:
		return 6
	case Rank9:
		return 9
	}
	return 0
}

type CardOption struct {
	// true -> +10, false -> -10
	Rank10Add bool `json:"rank_10_add"`

	// true -> +10, false -> -10
	RankQueenAdd bool `json:"rank_queen_add"`

	// RankKing will set Score to 99

	// next Player Id
	RankAceChangeNextPlayer uuid.UUID `json:"rank_ace_change_next_player"`

	// Rank8 will change Players order

	// draw from other Player
	RankJackDrawOneCardFromPlayer uuid.UUID `json:"rank_jack_draw_one_card_from_player"`

	// change all your Hand to other Player
	Rank7ChangeAllHandToPlayer uuid.UUID `json:"rank_7_change_all_hand_to_player"`
}
