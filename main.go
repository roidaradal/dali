package main

import (
	"fmt"
	"log"

	"github.com/roidaradal/dali/internal/dali"
)

func main() {
	node, err := dali.LoadNode()
	if err != nil {
		log.Fatal("Failed to get dali node:", err)
	}
	fmt.Println("Name:", node.Name)
	fmt.Println("IPAddr:", node.Addr)
}
