package gopen

import (
	"fmt"
	"sync"
	"time"
)

type App interface {
	Update()
}

func Run() {
	fmt.Println("Hello Gopen!")

	var wg sync.WaitGroup

	wg.Add(2)

	go logicLoop(&wg)
	go renderLoop(&wg)

	wg.Wait()

	fmt.Println("Test Loop Finished")
}

func logicLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 5; i++ {
		fmt.Println("Logic Loop Iteration:", i)
		time.Sleep(time.Second)
	}
}

func renderLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i < 5; i++ {
		fmt.Println("Render Loop Iteration:", i)
		time.Sleep(500 * time.Millisecond)
	}
}
