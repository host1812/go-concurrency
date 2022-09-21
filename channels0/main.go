package main

import (
	"fmt"
	"strings"
)

func say(ping <-chan string, pong chan<- string) {
	for {
		s, ok := <-ping
		if !ok {
			fmt.Println("some error getting from chan")
		}
		pong <- fmt.Sprintf("%s!!!", strings.ToUpper(s))
	}
}

func main() {
	ping := make(chan string)
	pong := make(chan string)
	go say(ping, pong)

	fmt.Println("enter what to say (q to quit):")
	for {
		fmt.Print("> ")
		var in string
		_, _ = fmt.Scanln(&in)
		if in == strings.ToLower("q") {
			break
		}
		ping <- in
		res := <-pong
		fmt.Println("said: ", res)
	}
	fmt.Println("finished")
	close(ping)
	close(pong)
}
