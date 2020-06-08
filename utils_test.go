package dl99

import (
	"fmt"
	"testing"
)

func TestADeckOfCards(t *testing.T) {
	cards := ADeckOfCards()
	if len(cards) != StandardNumbersOfPoker {
		t.Fatal()
	}

	cardsMap := make(map[Card]struct{}, len(cards))
	for _, card := range cards {
		cardsMap[card] = struct{}{}
	}
	if len(cardsMap) != StandardNumbersOfPoker {
		t.Fatal()
	}
}

func TestShuffle(t *testing.T) {
	cards := ADeckOfCards()
	Shuffle(cards)
	for _, card := range cards {
		fmt.Printf("%s\n", card.Name())
	}
}
