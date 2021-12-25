package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("Hello, world!")
	fmt.Printf("%s\n", os.Args[0])
	fmt.Println("Good bye, cruel world!")
	time.Sleep(2 * time.Second)
	fmt.Println("Good bye, cruel world!1")
	fmt.Println("Good bye, cruel world!2")
	fmt.Println("Good bye, cruel world!3")
}
