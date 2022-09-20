package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fatih/color"
)

const pizzasMax = 10

var pizzasMade, pizzasFailed, pizzasTotal int

type Producer struct {
	data chan PizzaOrder
	quit chan chan error
}

type PizzaOrder struct {
	pizzaId int
	message string
	success bool
}

func (p *Producer) Close() error {
	ch := make(chan error)
	p.quit <- ch
	return <-ch
}

func makePizza(pizzaNum int) *PizzaOrder {
	pizzaNum++
	if pizzaNum <= pizzasMax {
		delay := rand.Intn(5) + 1
		fmt.Printf("Received order %d\n", pizzaNum)
		rnd := rand.Intn(12) + 1
		msg := ""
		success := false

		if rnd < 5 {
			pizzasFailed++
		} else {
			pizzasMade++
		}
		pizzasTotal++

		fmt.Printf("Making pizza %d. It will take %ds...\n", pizzaNum, delay)
		time.Sleep(time.Duration(delay) * time.Second)

		if rnd <= 2 {
			msg = fmt.Sprintf("No ingredients (pizza: %d)", pizzaNum)
		} else if rnd <= 4 {
			msg = fmt.Sprintf("Cook drunk (pizza: %d)", pizzaNum)
		} else {
			success = true
			msg = fmt.Sprintf("Pizza done (pizza: %d)", pizzaNum)
		}

		p := PizzaOrder{
			pizzaId: pizzaNum,
			message: msg,
			success: success,
		}

		return &p
	}

	return &PizzaOrder{
		pizzaId: pizzaNum,
	}
}

func pizzeria(pizzaMaker *Producer) {
	var i = 0
	for {
		currentPizza := makePizza(i)
		if currentPizza != nil {
			i = currentPizza.pizzaId
			select {
			case pizzaMaker.data <- *currentPizza:
			case quit := <-pizzaMaker.quit:
				close(pizzaMaker.data)
				close(quit)
				return
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	color.Cyan("The Pizzeria is opened for business!")
	color.Cyan("------------------------------------")
	pizzaJob := &Producer{
		data: make(chan PizzaOrder),
		quit: make(chan chan error),
	}

	go pizzeria(pizzaJob)

	for p := range pizzaJob.data {
		if p.pizzaId <= pizzasMax {
			if p.success {
				color.Green(p.message)
				color.Green("Order %d out for delivery", p.pizzaId)
			} else {
				color.Red("Pizza failed, someones will be angry")
			}
		} else {
			color.Cyan("Done making pizzas...")
			err := pizzaJob.Close()
			if err != nil {
				color.Red("Error closing channel, err:", err)
			}
		}
	}
	color.Cyan("total: %d, failed: %d, made: %d", pizzasTotal, pizzasFailed, pizzasMade)
	switch {
	case pizzasFailed > 9:
		color.Red("bad")
	case pizzasFailed >= 6:
		color.Red("better, but still bad")
	default:
		color.Red("business as usual")
	}
}
