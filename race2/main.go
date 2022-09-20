package main

import (
	"log"
	"strconv"
	"sync"
)

var wg sync.WaitGroup

type Income struct {
	Source string
	Amount int
}

func main() {
	var balance int
	log.Printf("balance init: %d", balance)
	incomes := []Income{
		{Source: "base salary", Amount: 90},
		{Source: "gifts", Amount: 10},
		{Source: "part time", Amount: 20},
		{Source: "investments", Amount: 100},
		{Source: "2", Amount: 100},
		{Source: "3", Amount: 100},
		{Source: "4", Amount: 100},
	}

	for i := 0; i < 100000; i++ {
		incomes = append(incomes, Income{Source: strconv.Itoa(i), Amount: i})
	}
	var m sync.Mutex
	for i, in := range incomes {
		wg.Add(1)
		go func(i int, in Income, wg *sync.WaitGroup, m *sync.Mutex) {
			defer wg.Done()
			for w := 1; w < 52; w++ {
				m.Lock()
				balance = balance + in.Amount
				m.Unlock()
				// this is slower, because too many goroutines
				// wg.Add(1)
				// go func(i int, in Income, wg *sync.WaitGroup, m *sync.Mutex) {
				// 	defer wg.Done()
				// 	m.Lock()
				// 	balance = balance + in.Amount
				// 	m.Unlock()
				// }(i, in, wg, &m)
			}
		}(i, in, &wg, &m)
	}
	wg.Wait()
	log.Printf("balance final: %d", balance)
}
