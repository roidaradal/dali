package main

import (
	"fmt"
	"log"

	"github.com/roidaradal/dali/internal/dali"
	"github.com/roidaradal/fn/io"
)

func main() {
	command, options := io.GetCommandOptions(dali.HelpCmd)

	node, err := dali.LoadNode(command)
	if err != nil {
		log.Fatal("Failed to initialize: ", err)
	}
	fmt.Println(node)

	handler, ok := dali.CmdHandlers[command]
	if !ok {
		handler = dali.CmdHandlers[dali.HelpCmd]
	}
	err = handler(node, options)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
