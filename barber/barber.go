package main

import (
	"log"
	"time"

	"github.com/fatih/color"
)

type BarberShop struct {
	ShopCapacity    int
	HairCutDuration time.Duration
	NumberOfBarbers int
	BarbersDoneChan chan bool
	ClientsChan     chan string
	Opened          bool
}

func (b *BarberShop) AddBarber(barber string) {
	b.NumberOfBarbers++

	go func() {
		sleeping := false
		color.Yellow("%s: checks for clients", barber)

		for {
			if len(b.ClientsChan) == 0 {
				color.Yellow("%s: no clients, taking a sleep", barber)
				sleeping = true
			}
			c, ok := <-b.ClientsChan
			if ok {
				if sleeping {
					color.Yellow("%s: waked by %s", barber, c)
					sleeping = false
				}
				b.CutHair(barber, c)
			} else {
				b.CloseBarber(barber)
				return
			}
		}
	}()
}

func (b *BarberShop) CutHair(barber, c string) {
	log.Printf("%s: cutting hair for %s", barber, c)
	time.Sleep(b.HairCutDuration)
	log.Printf("%s: done hair for %s", barber, c)
}

func (b *BarberShop) CloseBarber(barber string) {
	color.Cyan("%s: going home", barber)
	b.BarbersDoneChan <- true
}

func (b *BarberShop) Close() {
	color.Cyan("Closing")
	close(b.ClientsChan)
	b.Opened = false

	for a := 1; a <= b.NumberOfBarbers; a++ {
		<-b.BarbersDoneChan
	}
	close(b.BarbersDoneChan)
	color.Green("Closed")
}

func (b *BarberShop) AddClient(name string) {
	color.Green("%s: arrived", name)
	if b.Opened {
		select {
		case b.ClientsChan <- name:
			color.Yellow("%s: took seat in waiting room", name)
		default:
			color.Red("%s: leaves (waiting room is full", name)
		}
	} else {
		color.Red("%s: leaves (shop is closed)", name)
	}
}
