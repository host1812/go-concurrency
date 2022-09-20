package main

import (
	"fmt"
	"sync"
)

var msg string
var wg sync.WaitGroup

func update(s string, wg *sync.WaitGroup, m *sync.Mutex) {
	defer wg.Done()
	m.Lock()
	msg = s
	m.Unlock()
}

func update2(s string, wg *sync.WaitGroup) {
	defer wg.Done()
	msg = s
}

func main() {
	msg = "hello"
	var m sync.Mutex
	wg.Add(2)
	go update("do", &wg, &m)
	go update("oh", &wg, &m)
	go update2("do", &wg)
	go update2("oh", &wg)
	wg.Wait()

	fmt.Println(msg)
}
