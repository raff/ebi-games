package main

import (
	//"fmt"
	"math/rand"
)

const (
	ndups = 8  // 8 copies of each symbol (and number of symbols per card)
	nsyms = 57 // total number of symbols (and number of cards)
)

func print(c []int) {
	//fmt.Printf("%2v\n", c)
}

func getCards() (cards [][]int) {
	N := ndups - 1

	addcard := func(c []int) {
		rand.Shuffle(len(c), func(i, j int) {
			c[i], c[j] = c[j], c[i]
		})

		cards = append(cards, c)
	}

	// Fist card
	var card []int

	for i := 0; i <= N; i++ {
		card = append(card, i+1)
	}

	addcard(card)

	// N following cards
	for i := 0; i < N; i++ {
		var card []int

		card = append(card, 1)

		for j := 0; j < N; j++ {
			card = append(card, (N+1)+(N*i)+j)
		}

		addcard(card)
	}

	// N*N following cards
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			var card []int

			card = append(card, i+1)

			for k := 0; k < N; k++ {
				card = append(card, (N+1)+(N*k)+(i*k+j)%N)
			}

			addcard(card)
		}
	}

	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})

	return
}

func match(c1, c2 []int) int {
	for _, v1 := range c1 {
		for _, v2 := range c2 {
			if v1 == v2 {
				return v1
			}
		}
	}

	return 0
}

/*
func main() {
	cards := getCards()

	fmt.Println()

	for _, c := range cards {
		print(c)
	}

	fmt.Println()
	fmt.Println(len(cards), "cards")
	fmt.Println()

	for i := 1; i < len(cards); i++ {
		c1, c2 := cards[i-1], cards[i]
		fmt.Printf("%2v %2v -> %v\n", c1, c2, match(c1, c2))
	}
}

*/
