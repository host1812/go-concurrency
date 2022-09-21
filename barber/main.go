package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

var capacity = 10
var rate = 100000
var duration = 1 * time.Second
var open = 10 * time.Second

func main() {
	rand.Seed(time.Now().UnixNano())
	color.Yellow("Barber")
	color.Yellow("------")

	clients := make(chan string, capacity)
	done := make(chan bool)

	shop := BarberShop{
		ShopCapacity:    capacity,
		HairCutDuration: duration,
		NumberOfBarbers: 0,
		ClientsChan:     clients,
		BarbersDoneChan: done,
		Opened:          true,
	}

	color.Green("Opened")
	shop.AddBarber("A")

	shopCh := make(chan bool)
	closed := make(chan bool)
	go func() {
		<-time.After(open)
		shopCh <- true
		shop.Close()
		closed <- true
	}()

	i := 1
	go func() {
		for {
			r := rand.Int() % (2 * rate)
			select {
			case <-shopCh:
				return
			case <-time.After(time.Microsecond * time.Duration(r)):
				shop.AddClient(fmt.Sprintf("client #%d", i))
				i++
			}
		}
	}()
	<-closed
}
