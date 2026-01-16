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
	handler, ok := dali.CmdHandlers[command]
	if !ok {
		handler = dali.CmdHandlers[dali.HelpCmd]
	}
	err = handler(node, options)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Get command and options
func getCommandArgs() (string, dict.StringMap) {
	args := os.Args[1:]
	if len(args) < 1 {
		args = []string{dali.HelpCmd}
	}
	command := strings.ToLower(args[0])
	options := make(dict.StringMap)
	for _, pair := range args[1:] {
		parts := strings.Split(pair, "=")
		key := strings.ToLower(parts[0])
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		options[key] = value
	}
	return command, options
}
