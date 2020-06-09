package dl99

import (
	"fmt"
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
	default:
		panic("invalid Suit Name")
	}
}

type Rank uint8

const (
	NoRank Rank = iota
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
	case NoRank:
		return "No Rank"
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
	default:
		panic("invalid Rank Name")
	}
}

// 1 -> 13 Diamond
// 14 -> 26 Club
// 27 -> 39 Heart
// 40 -> 52 Spade
// 53 -> BlackJoker
// 54 -> RedJoker
type Card uint8

func (card Card) Suit() Suit {
	if 1 <= card && card <= 13 {
		return Diamond
	} else if 14 <= card && card <= 26 {
		return Club
	} else if 27 <= card && card <= 39 {
		return Heart
	} else if 40 <= card && card <= 52 {
		return Spade
	} else if card == 53 {
		return BlackJoker
	} else if card == 54 {
		return RedJoker
	} else {
		panic(fmt.Sprintf("invalid card suit for: %d", card))
	}
}

func (card Card) Rank() Rank {
	if 1 <= card && card <= 52 {
		switch card % 13 {
		case 1:
			return RankAce
		case 2:
			return Rank2
		case 3:
			return Rank3
		case 4:
			return Rank4
		case 5:
			return Rank5
		case 6:
			return Rank6
		case 7:
			return Rank7
		case 8:
			return Rank8
		case 9:
			return Rank9
		case 10:
			return Rank10
		case 11:
			return RankJack
		case 12:
			return RankQueen
		case 0:
			return RankKing
		}
	}
	return NoRank
}

func (card Card) Name() string {
	suit := card.Suit()
	switch suit {
	case RedJoker, BlackJoker:
		return suit.Name()
	default:
		return fmt.Sprintf("%s of %s", card.Rank().Name(), suit.Name())
	}
}

func (card Card) Score() int {
	switch card.Rank() {
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
	RankAceChangeNextPlayer string `json:"rank_ace_change_next_player"`

	// Rank8 will change Players order

	// draw from other Player
	RankJackDrawOneCardFromPlayer string `json:"rank_jack_draw_one_card_from_player"`

	// change all your hand to other Player
	Rank7ChangeAllHandToPlayer string `json:"rank_7_change_all_hand_to_player"`
}

var (
	deck = []Card{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
		27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
		40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
		53, 54,
	}

	freeBattleDeadline99Deck = []Card{
		1,  /*2,*/ 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, /*15,*/ 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
		27, /*28,*/ 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
		40, /*41,*/ 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
		/* 53, 54, */
	}
)
