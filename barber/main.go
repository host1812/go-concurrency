package main

import (
	"math/rand"
	"time"

	"github.com/fatih/color"
)

var capacity = 10
var rate = 100
var duration = 1000 * time.Microsecond
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
		Open:            true,
	}

	color.Green("Opened")
	shop.AddBarber("A")
	time.Sleep(5 * time.Second)
}
