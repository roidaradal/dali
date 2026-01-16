// Package dali contains implementations of the dali tool
package dali

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
)

const currentVersion string = "0.1.2"

const (
	HelpCmd string = "help"
	openCmd string = "open"
)

var CmdHandlers = map[string]func(*Node, dict.StringMap) error{
	HelpCmd:   cmdHelp,
	"version": cmdVersion,
	"set":     cmdSet,
	"find":    cmdFind,
	openCmd:   cmdOpen,
	"send":    cmdSend,
}

// List of commands, ordered for help
var commands = []string{"set", openCmd, "find", "send", "version", HelpCmd}

// Load user node
func LoadNode() (*Node, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, wrapErr("cannot load home dir", err)
	}

	var cfg *Config = nil

	// Initialize dali config file, if it does not exist
	path := filepath.Join(homeDir, cfgPath)
	if !io.PathExists(path) {
		hostName, err := os.Hostname()
		if err != nil {
			return nil, wrapErr("failed to get hostname", err)
		}
		hostName = compressName(hostName)
		cfg = newConfig(hostName)
		cfg.Path = path
		err = cfg.Save()
		if err != nil {
			return nil, wrapErr("failed to initialize dali config", err)
		}
	}

	if cfg == nil {
		// Load dali config file
		cfg, err = io.ReadJSON[Config](path)
		if err != nil {
			return nil, wrapErr("failed to load dali config", err)
		}
		cfg.Path = path
	}

	// Get local IP address
	addr, err := getLocalIPv4Address()
	if err != nil {
		return nil, wrapErr("failed to get local IP addr", err)
	}

	node := &Node{
		Config: cfg,
		Addr:   addr,
	}
	return node, nil
}

// Help command handler
func cmdHelp(node *Node, options dict.StringMap) error {
	fmt.Println("\nUsage: dali <command> (option=value)*")
	fmt.Printf("\n%d dali commands:\n", len(commands))
	for _, cmd := range commands {
		fmt.Printf("  • %s\n", cmd)
	}
	return nil
}

// Version command handler
func cmdVersion(node *Node, options dict.StringMap) error {
	fmt.Printf("\ndali v%s\n", currentVersion)
	return nil
}

// Set command handler
func cmdSet(node *Node, options dict.StringMap) error {
	// Options: name=NAME, timeout=X, wait=X
	cfg := node.Config
	for k, v := range options {
		switch k {
		case "name":
			// Make sure name has no spaces
			cfg.Name = compressName(v)
		case "timeout", "wait":
			// Clip new timeout value, with floor = minTimeout
			cfg.Timeout = max(minTimeout, number.ParseInt(v))
		}
	}
	if err := cfg.Save(); err != nil {
		return err
	}
	fmt.Println("Updated:")
	fmt.Println(node)
	return nil
}

// Find command handler
func cmdFind(node *Node, options dict.StringMap) error {
	// TODO: add option name=NAME, ip=IPAddr
	fmt.Println(findingMessage(node))
	peers, err := discoverPeers(time.Duration(node.Timeout) * time.Second)
	if err != nil {
		return err
	}

	if len(peers) == 0 {
		fmt.Println("No peers found.")
		return nil
	}

	fmt.Printf("Found %d peers:\n", len(peers))
	maxLength := maxPeerNameLength(peers)
	template := fmt.Sprintf("  • %%%ds : %%s\n", maxLength)
	for _, peer := range peers {
		fmt.Printf(template, peer.Name, peer.Addr)
	}
	return nil
}

// Open command handler
func cmdOpen(node *Node, options dict.StringMap) error {
	// Options: port=CUSTOM_PORT, output=OUT_DIR, out=OUT_DIR
	listenPort := transferPort // default port
	outputDir := "."           // default: current dir
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
		runDiscoveryListener(node.Name, listenPort)
	}()

	err := receiveFiles(listenPort, outputDir)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return nil
}

// Send command handler
func cmdSend(node *Node, options dict.StringMap) error {

	// Options: file=FILE_PATH, to=IPADDR:PORT, for=NAME
	filePath, peerAddr, peerName := "", "", anyone
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
		return fmt.Errorf("missing file path. Use file=<filePath>")
	}

	if !io.PathExists(filePath) {
		return fmt.Errorf("file %q does not exist", filePath)
	}

	if peerAddr == "" {
		// Find peers if no set peer address
		fmt.Println(findingMessage(node))
		peers, err := discoverPeers(time.Duration(node.Timeout) * time.Second)
		if err != nil {
			return wrapErr("discovery failed", err)
		}

		if len(peers) == 0 {
			fmt.Printf("No peers found. Make sure another device is running `dali %s`\n", openCmd)
			return nil
		}

		autoSelect := false
		if peerName != anyone {
			autoSelect = true
			targetName := strings.ToLower(peerName)
			peers = list.Filter(peers, func(peer Peer) bool {
				return strings.ToLower(peer.Name) == targetName
			})
			if len(peers) == 0 {
				fmt.Printf("Peer %q not found. Make sure device is running `dali %s`\n", peerName, openCmd)
				return nil
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
				return fmt.Errorf("invalid selection")
			}
			peerIdx = choice - 1
		}

		peer := peers[peerIdx]
		peerName, peerAddr = peer.Name, peer.Addr
	}

	fmt.Printf("Sending %q to %s (%s)...", filePath, peerName, peerAddr)
	return sendFile(peerAddr, filePath)
}
