package main

import (
	"fmt"
	"time"
)

func listen(ch chan int) {
	for {
		i := <-ch
		fmt.Println("got:", i)

		// do something
		time.Sleep(1 * time.Second)
	}
}

func main() {
	ch := make(chan int, 10)
	go listen(ch)

	for i := 0; i < 100; i++ {
		fmt.Println("sending:", i)
		ch <- i
		fmt.Println("sent:", i)
	}
	fmt.Println("done")
	close(ch)
}
