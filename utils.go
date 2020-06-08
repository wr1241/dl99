package dl99

import (
	"math/rand"
	"time"
)

const (
	StandardNumbersOfPoker = 13*4 + 2
)

func ADeckOfCards(ignoredCards ...Card) []Card {
	ignoredCardsMap := make(map[Card]struct{})
	for _, card := range ignoredCards {
		ignoredCardsMap[card] = struct{}{}
	}

	cards := make([]Card, 0, StandardNumbersOfPoker)
	for _, suit := range []Suit{Heart, Diamond, Club, Spade} {
		for _, rank := range []Rank{
			RankAce,
			Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, Rank10,
			RankJack, RankQueen, RankKing,
		} {
			card := Card{
				Suit: suit,
				Rank: rank,
			}
			if _, ok := ignoredCardsMap[card]; ok {
				continue
			}
			cards = append(cards, card)
		}
	}
	cards = append(cards, Card{Suit: RedJoker}, Card{Suit: BlackJoker})
	return cards
}

// Fisher-Yates
func Shuffle(cards []Card) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var j int
	for i := 0; i < len(cards); i++ {
		j = rng.Intn(len(cards)-i) + i
		cards[i], cards[j] = cards[j], cards[i]
	}
}
