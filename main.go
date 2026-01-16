package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/roidaradal/dali/internal/dali"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
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
		// Options: name=NAME, timeout=X, wait=X
		err = node.Config.Update(options)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Updated:")
		fmt.Println(node)
	case "find":
		// TODO: add option name=NAME, ip=IPAddr
		fmt.Printf("Finding peers on local network for %ds...\n", node.Timeout)
		peers, err := dali.DiscoverPeers(time.Duration(node.Timeout) * time.Second)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(peers) == 0 {
			fmt.Println("No peers found.")
			return
		}

		fmt.Printf("Found %d peers:\n", len(peers))
		maxLength := slices.Max(list.Map(peers, func(p dali.Peer) int {
			return len(p.Name)
		}))
		template := fmt.Sprintf("  • %%%ds : %%s\n", maxLength)
		for _, peer := range peers {
			fmt.Printf(template, peer.Name, peer.Addr)
		}
	case "open":
		// Options: port=CUSTOM_PORT, output=OUT_DIR, out=OUT_DIR
		listenPort := dali.TransferPort // default port
		outputDir := "."                // default: current dir
		for k, v := range options {
			switch k {
			case "port":
				customPort := number.ParseInt(v)
				if customPort > 0 {
					listenPort = uint16(customPort)
				}
			case "output", "out":
				outputDir = v
			}
		}
		fmt.Printf("Output folder: %s\n", outputDir)
		fmt.Printf("Listening for requests on local network at port %d...\n", listenPort)
		dali.RunDiscoveryListener(node.Name, listenPort)
	default:
		fmt.Println("\nUsage: dali <command> (option=value...)")
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
