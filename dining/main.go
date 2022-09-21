package main

import (
	"fmt"
	"sync"
	"time"
)

type Philosopher struct {
	Name        string
	RightForkId int
	LeftForkId  int
}

var Philosophers = []Philosopher{
	{Name: "A", LeftForkId: 4, RightForkId: 0},
	{Name: "B", LeftForkId: 0, RightForkId: 1},
	{Name: "C", LeftForkId: 1, RightForkId: 2},
	{Name: "D", LeftForkId: 2, RightForkId: 3},
	{Name: "E", LeftForkId: 3, RightForkId: 4},
}

var hunger = 3
var eatTime = 0 * time.Second
var thinkTime = 0 * time.Second
var sleepTime = 0 * time.Second

func DiningProblem(
	ph Philosopher,
	wg *sync.WaitGroup,
	forks map[int]*sync.Mutex,
	seated *sync.WaitGroup,
	completed *[]Philosopher,
	cm *sync.Mutex,
) {
	defer wg.Done()
	fmt.Printf("%s: seated at the table\n", ph.Name)
	seated.Done()

	seated.Wait()

	for i := hunger; i > 0; i-- {
		if ph.LeftForkId > ph.RightForkId {
			forks[ph.LeftForkId].Lock()
			fmt.Printf("\t%s: took left fork\n", ph.Name)
			forks[ph.RightForkId].Lock()
			fmt.Printf("\t%s: took right fork\n", ph.Name)
		} else {
			forks[ph.RightForkId].Lock()
			fmt.Printf("\t%s: took right fork\n", ph.Name)
			forks[ph.LeftForkId].Lock()
			fmt.Printf("\t%s: took left fork\n", ph.Name)
		}
		fmt.Printf("\t%s: eating plate #%d\n", ph.Name, i)
		time.Sleep(eatTime)
		fmt.Printf("\t%s: thinking\n", ph.Name)
		time.Sleep(thinkTime)
		fmt.Printf("\t%s: sleeping\n", ph.Name)
		time.Sleep(sleepTime)
		forks[ph.LeftForkId].Unlock()
		forks[ph.RightForkId].Unlock()
		fmt.Printf("\t%s: put down forks\n", ph.Name)
	}
	fmt.Printf("%s: satisfied and left table\n", ph.Name)
	cm.Lock()
	*completed = append(*completed, ph)
	cm.Unlock()
}

func Dine() {
	wg := &sync.WaitGroup{}
	wg.Add(len(Philosophers))

	seated := &sync.WaitGroup{}
	seated.Add(len(Philosophers))

	var forks = make(map[int]*sync.Mutex)
	var cm = &sync.Mutex{}

	var completed []Philosopher
	for i := 0; i < len(Philosophers); i++ {
		forks[i] = &sync.Mutex{}
	}

	for i := 0; i < len(Philosophers); i++ {
		go DiningProblem(Philosophers[i], wg, forks, seated, &completed, cm)
	}

	wg.Wait()
	fmt.Println("order everyone complete:", completed)
}

func main() {
	fmt.Println("Started")
	fmt.Println("Table is empty")
	Dine()
	fmt.Println("Table is empty")
	fmt.Println("Finished")
}
