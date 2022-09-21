package main

import (
	"fmt"
	"time"
)

func server1(ch chan string) {
	for {
		time.Sleep(6 * time.Second)
		ch <- "this is from server1"
	}
}

func server2(ch chan string) {
	for {
		time.Sleep(3 * time.Second)
		ch <- "this is from server2"
	}
}

func main() {
	fmt.Println("started")
	ch1 := make(chan string)
	ch2 := make(chan string)
	go server1(ch1)
	go server2(ch2)
	for {
		select {
		case s1 := <-ch1:
			fmt.Println("case one:", s1)
		case s2 := <-ch1:
			fmt.Println("case two:", s2)
		case s3 := <-ch2:
			fmt.Println("case three:", s3)
		case s4 := <-ch2:
			fmt.Println("case four:", s4)
		default:
		}
	}
	fmt.Println("finished")
}
