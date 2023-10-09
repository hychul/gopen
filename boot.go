package gopen

import "fmt"

type App interface {
	Update()
}

func Run() {
	fmt.Println("Hello Gopen!")
}
