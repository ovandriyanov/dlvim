package main

import (
	"fmt"
	"os"
	"time"
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
	fmt.Println("Hello, world!")
	fmt.Println(k)
	fmt.Printf("%s\n", os.Args[0])
	fmt.Println("Good bye, cruel world!")
	time.Sleep(2 * time.Second)
	fmt.Println("Good bye, cruel world!1")
	fmt.Println("Good bye, cruel world!2")
	fmt.Println("Good bye, cruel world!3")
}
