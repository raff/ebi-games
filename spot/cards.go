package main

import (
// "fmt"
// "math/rand"
)

const (
	ndups = 8  // 8 copies of each symbol (and number of symbols per card)
	nsyms = 57 // total number of symbols (and number of cards)
)

func getCards() (cards [][]int) {
	var syms []int

	for r := 0; r < ndups; r++ {
		s := r + 1

		for t := 0; t < nsyms; t++ {
			syms = append(syms, s)
			s = (s % nsyms) + 1
		}
	}

	getsym := func() (s int) {
		s, syms = syms[0], syms[1:]
		return
	}

	for k := 0; k < ndups; k++ {
		first := getsym()

		for r := 0; r < ndups; r++ {
			card := []int{first}

			for k := 1; k < ndups; k++ {
				card = append(card, getsym())
			}

			//rand.Shuffle(len(card), func(i, j int) {
			//	card[i], card[j] = card[j], card[i]
			//})

			cards = append(cards, card)
		}
	}

	//rand.Shuffle(len(cards), func(i, j int) {
	//	cards[i], cards[j] = cards[j], cards[i]
	//})

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

	for _, c := range cards {
		fmt.Printf("%2v\n", c)
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
