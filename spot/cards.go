package main

const (
	ndups = 8  // 8 copies of each object (and number of objects per card)
	nsyms = 57 // total number of objects (and number of cards)
)

func getCards() (cards [][]int) {
	var syms []int

	for r := 0; r < ndups; r++ {
		s := r + 1

		for t := 0; t < nsyms; t++ {
			syms = append(syms, s)
			s++

			if s > nsyms {
				s = 1
			}
		}
	}

	getsym := func() (s int) {
		s, syms = syms[0], syms[1:]
		return
	}

	for k := 0; k < ndups; k++ {
		card := []int{getsym()}

		for r := 0; r < ndups; r++ {
			for k := 1; k < ndups; k++ {
				card = append(card, getsym())
			}

			fmt.Println(card)
			cards = append(cards, card)
			card = []int{card[0]}
		}
	}

	return
}
