package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("CharLine Client")
		fmt.Println("Usage: client <command> [args]")
		fmt.Println("Commands:")
		fmt.Println("  hello    - Print hello message")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "hello":
		fmt.Println("Hello from CharLine Client!")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
