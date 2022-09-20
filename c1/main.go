package main

import (
	"fmt"
	"log"
	"sync"
)

func some(s string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println("some: started")
	log.Println("some: s:", s)
	log.Println("some: finished")
}

func main() {
	log.Println("main: started")
	var wg sync.WaitGroup
	words := []string{
		"alpha",
		"beta",
		"delta",
		"zeta",
		"er",
		"df",
		"w311",
	}
	for i, w := range words {
		wg.Add(1)
		go some(fmt.Sprintf("%d - %s", i, w), &wg)
	}

	wg.Wait()

	log.Println("main: finished")
}
