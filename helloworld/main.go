package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, world!")
	fmt.Printf("%s\n", os.Args[0])
	fmt.Println("Good bye, cruel world!")
}
