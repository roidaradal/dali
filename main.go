package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/roidaradal/dali/internal/dali"
	"github.com/roidaradal/fn/dict"
)

func main() {
	node, err := dali.LoadNode()
	if err != nil {
		log.Fatal("Failed to get dali node:", err)
	}
	fmt.Println(node)

	command, options := getCommandArgs()
	switch command {
	case "set":
		err = node.Config.Update(options)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Updated:")
		fmt.Println(node)
	default:
		fmt.Println("Usage: dali <command> (option=value...)")
	}
}

func getCommandArgs() (string, dict.StringMap) {
	args := os.Args[1:]
	if len(args) < 1 {
		args = []string{"help"}
	}
	command := args[0]
	options := make(dict.StringMap)
	for _, pair := range args[1:] {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			continue
		}
		options[parts[0]] = parts[1]
	}
	return command, options
}
