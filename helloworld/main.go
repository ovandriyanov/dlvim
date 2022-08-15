package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/xerrors"
)

type kek struct {
	i int
	s string
}

func main() {
	k := kek{
		i: 1,
		s: "xyz",
	}
	err := xerrors.Errorf("kek")
	fmt.Println(err)
	fmt.Println("Hello, world!")
	fmt.Println(k)
	done := make(chan struct{})
	go func() {
		done <- struct{}{}
	}()
	fmt.Printf("%s\n", os.Args[0])
	fmt.Println("Good bye, cruel world!")
	<-done
	time.Sleep(2 * time.Second)
	fmt.Println("Good bye, cruel world!1")
	fmt.Println("Good bye, cruel world!2")
	fmt.Println("Good bye, cruel world!3")
}
