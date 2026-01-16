package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/roidaradal/dali/internal/dali"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

func main() {
	node, err := dali.LoadNode()
	if err != nil {
		log.Fatal("Failed to get dali node:", err)
	}
	fmt.Println(node)
	findMessage := fmt.Sprintf("Finding peers on local network for %ds...\n", node.Timeout)
	findTimeout := time.Duration(node.Timeout) * time.Second

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
		fmt.Println(findMessage)
		peers, err := dali.DiscoverPeers(findTimeout)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(peers) == 0 {
			fmt.Println("No peers found.")
			return
		}

		fmt.Printf("Found %d peers:\n", len(peers))
		maxLength := maxPeerNameLength(peers)
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

		// Run discovery listener in the background
		go func() {
			dali.RunDiscoveryListener(node.Name, listenPort)
		}()

		err := dali.ReceiveFiles(listenPort, outputDir)
		if err != nil {
			fmt.Println("Error:", err)
		}
	case "send":
		// Options: file=FILE_PATH, to=IPADDR:PORT, for=NAME
		filePath, peerAddr, peerName := "", "", dali.Any
		for k, v := range options {
			switch k {
			case "file":
				filePath = v
			case "to":
				peerAddr = v
			case "for":
				peerName = v
			}
		}

		if filePath == "" {
			fmt.Println("Error: missing file path. Use file=<filePath>")
			return
		}

		if !io.PathExists(filePath) {
			fmt.Printf("Error: file %q does not exist\n", filePath)
			return
		}

		if peerAddr == "" {
			// Find peers if no set peer address
			fmt.Println(findMessage)
			peers, err := dali.DiscoverPeers(findTimeout)
			if err != nil {
				fmt.Println("Discovery failed:", err)
				return
			}

			if len(peers) == 0 {
				fmt.Println("No peers found. Make sure another device is running `dali open`")
				return
			}

			autoSelect := false
			if peerName != dali.Any {
				autoSelect = true
				targetName := strings.ToLower(peerName)
				peers = list.Filter(peers, func(peer dali.Peer) bool {
					return strings.ToLower(peer.Name) == targetName
				})
				if len(peers) == 0 {
					fmt.Printf("Peer %q not found. Make sure device is running `dali open`\n", peerName)
					return
				}
			}

			var peerIdx int
			if autoSelect && len(peers) == 1 {
				peerIdx = 0
			} else {
				// Let user select recipient
				numPeers := len(peers)
				fmt.Printf("\nFound %d peers:\n", numPeers)
				maxLength := maxPeerNameLength(peers)
				template := fmt.Sprintf("  [%%2d] %%%ds : %%s\n", maxLength)
				for i, peer := range peers {
					fmt.Printf(template, i+1, peer.Name, peer.Addr)
				}

				fmt.Println("\nEnter peer number to send to:")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				choice := number.ParseInt(input)
				if choice < 1 || choice > numPeers {
					fmt.Println("Error: invalid selection")
					return
				}
				peerIdx = choice - 1
			}

			peer := peers[peerIdx]
			peerName, peerAddr = peer.Name, peer.Addr
		}

		fmt.Printf("Sending %q to %s (%s)...", filePath, peerName, peerAddr)
		// TODO: dali.SendFile(peerAddr, filePath)
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

func maxPeerNameLength(peers []dali.Peer) int {
	return slices.Max(list.Map(peers, func(p dali.Peer) int {
		return len(p.Name)
	}))
}
