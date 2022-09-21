package main

import (
	"time"

	"github.com/fatih/color"
)

type BarberShop struct {
	ShopCapacity    int
	HairCutDuration time.Duration
	NumberOfBarbers int
	BarbersDoneChan chan bool
	ClientsChan     chan string
	Open            bool
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
				b.Close(barber)
				return
			}
		}
	}()
}

func (b *BarberShop) CutHair(barber, c string) {
	color.Green("%s: cutting hair for %s", barber, c)
	time.Sleep(b.HairCutDuration)
	color.Green("%s: done hair for %s", barber, c)
}

func (b *BarberShop) Close(barber string) {
	color.Cyan("%s: going home", b)
	b.BarbersDoneChan <- true
}
