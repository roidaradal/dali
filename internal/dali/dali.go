// Package dali contains implementations of the dali tool
package dali

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
)

const currentVersion string = "0.1.3"

const (
	HelpCmd    string = "help"
	openCmd    string = "open"
	versionCmd string = "version"
	setCmd     string = "set"
	findCmd    string = "find"
	sendCmd    string = "send"
	updateCmd  string = "update"
	logsCmd    string = "logs"
)

var CmdHandlers = map[string]func(*Node, dict.StringMap) error{
	HelpCmd:    cmdHelp,
	versionCmd: cmdVersion,
	setCmd:     cmdSet,
	findCmd:    cmdFind,
	openCmd:    cmdOpen,
	sendCmd:    cmdSend,
	updateCmd:  cmdUpdate,
	logsCmd:    cmdLogs,
}

// List of commands, ordered for help
var commands = []string{setCmd, openCmd, sendCmd, findCmd, updateCmd, logsCmd, versionCmd, HelpCmd}

var cmdColor = map[string]func(string) string{
	HelpCmd:    str.Yellow,
	versionCmd: str.Red,
	setCmd:     str.Red,
	findCmd:    str.Cyan,
	openCmd:    str.Yellow,
	sendCmd:    str.Green,
	updateCmd:  str.Blue,
	logsCmd:    str.Violet,
}

var cmdText = dict.StringMap{
	HelpCmd:    "display help message",
	versionCmd: "display current version",
	setCmd:     "update name and waiting time",
	findCmd:    "discover open machines on local network",
	openCmd:    "opens the machine to receive files and discovery",
	sendCmd:    "send file to an open machine",
	updateCmd:  "update dali to latest (or specific) version",
	logsCmd:    "view activity logs",
}

var cmdOptions = map[string][][2]string{
	setCmd: {
		{"name={NAME}", "set your name (no spaces)"},
		{"wait={TIMEOUT_SECS}", "set waiting time (in seconds) for finding peers"},
		{"timeout={TIMEOUT_SECS}", "set waiting time (in seconds) for finding peers"},
	},
	openCmd: {
		{"", "listen on default port (45679)"},
		{"port={PORT}", "listen on custom port"},
		{"out={OUT_DIR}", "set custom output folder"},
		{"output={OUT_DIR}", "set custom output folder"},
		{"accept=auto", "auto-accepts incoming file transfers"},
		{"overwrite", "overwrite old file path if it exists"},
	},
	sendCmd: {
		{"file={FILE_PATH}", "finds peers and select one to send file to"},
		{"file={FILE_PATH} for={NAME}", "find {NAME} peer and send file"},
		{"file={FILE_PATH} to={IPADDR:PORT}", "send file to specific address in local network"},
		{"file={FILE_PATH} auto=1", "send file automatically if only 1 peer found"},
	},
	findCmd: {
		{"", "look for all peers in local network"},
		{"name={NAME}", "look for peer {NAME} in local network"},
		{"ip={IP_ADDR}", "look for peer with specified IP address in local network"},
	},
	updateCmd: {
		{"", "update to latest version"},
		{"v=0.1.0", "update to specific version"},
		{"version=0.1.0", "update to specific version"},
	},
	logsCmd: {
		{"", "view all logs"},
	},
}

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
		// Prompt user for custom name and timeout
		fmt.Println("Welcome to dali!")
		fmt.Print("Enter your name: ")
		name := readInput()
		if name != "" {
			cfg.Name = name
		}
		fmt.Print("Set timeout (default: 3s): ")
		wait := number.ParseInt(readInput())
		if wait > 0 {
			cfg.Timeout = wait
		}
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
func cmdHelp(_ *Node, _ dict.StringMap) error {
	fmt.Println("\nUsage: dali <command> (option=value)*")
	fmt.Printf("\n%d dali commands:\n", len(commands))
	for _, cmd := range commands {
		// Get command description
		description := ""
		if text, ok := cmdText[cmd]; ok {
			description = fmt.Sprintf("- %s", text)
		}

		// Get command color
		coloredCmd := cmd
		if color, ok := cmdColor[cmd]; ok {
			coloredCmd = color(cmd)
		}

		fmt.Printf("  • %s %s\n", coloredCmd, description)

		// Get command options
		optionPairs := cmdOptions[cmd]
		if len(optionPairs) == 0 {
			continue
		}
		descriptions := list.Map(optionPairs, func(pair [2]string) string {
			return pair[1]
		})
		options := list.Map(optionPairs, func(pair [2]string) string {
			return fmt.Sprintf("dali %s %s", coloredCmd, pair[0])
		})
		maxLength := slices.Max(list.Map(options, str.Length))
		template := fmt.Sprintf("      ▪ %%-%ds - %%s\n", maxLength)
		for i, option := range options {
			fmt.Printf(template, option, descriptions[i])
		}
		fmt.Println()
	}
	return nil
}

// Version command handler
func cmdVersion(_ *Node, _ dict.StringMap) error {
	version := fmt.Sprintf("dali v%s", currentVersion)
	fmt.Printf("\n%s\n", str.Green(version))
	return nil
}

// Update command handler
func cmdUpdate(_ *Node, options dict.StringMap) error {
	version := "latest"
	for k, v := range options {
		switch k {
		case "v", "version":
			version = v
		}
	}
	cmd1, cmd2 := "go", "install"
	cmd3 := fmt.Sprintf("github.com/roidaradal/dali@%s", version)

	fmt.Printf("Running: %s %s %s ... ", cmd1, cmd2, cmd3)
	cmd := exec.Command("cmd", "/c", cmd1, cmd2, cmd3)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("FAIL")
		return err
	}
	fmt.Println("OK")
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
	// Options: name=NAME, ip=IPAddr
	peerName, peerAddr := anyone, anyone
	for k, v := range options {
		switch k {
		case "name":
			// Make sure name has no spaces
			peerName = compressName(v)
		case "ip":
			peerAddr = v
		}
	}

	fmt.Println(findingMessage(node))
	peers, err := discoverPeers(time.Duration(node.Timeout)*time.Second, Peer{Name: peerName, Addr: peerAddr})
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
	// Options: port=CUSTOM_PORT, output=OUT_DIR, out=OUT_DIR, accept=auto, overwrite
	listenPort := transferPort // default port
	outputDir := "."           // default: current dir
	autoAccept, overwrite := false, false
	for k, v := range options {
		switch k {
		case "port":
			customPort := number.ParseInt(v)
			if customPort > 0 {
				listenPort = uint16(customPort)
			}
		case "output", "out":
			outputDir = v
		case "accept":
			autoAccept = strings.ToLower(v) == "auto"
		case "overwrite":
			overwrite = true
		}
	}
	fmt.Printf("Output folder: %s\n", outputDir)
	fmt.Printf("Listening for requests on local network at port %d...\n", listenPort)

	// Run discovery listener in the background
	go func() {
		runDiscoveryListener(node.Name, listenPort)
	}()

	err := receiveFiles(node, listenPort, outputDir, autoAccept, overwrite)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return nil
}

// Send command handler
func cmdSend(node *Node, options dict.StringMap) error {
	// Options: file=FILE_PATH, to=IPADDR:PORT, for=NAME, auto=1
	filePath, peerAddr, peerName := "", "", anyone
	autoSend := false
	for k, v := range options {
		switch k {
		case "file":
			filePath = v
		case "to":
			peerAddr = v
		case "for":
			peerName = v
		case "auto":
			autoSend = v == "1"
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
		peers, err := discoverPeers(time.Duration(node.Timeout)*time.Second, Peer{Name: peerName, Addr: anyone})
		if err != nil {
			return wrapErr("discovery failed", err)
		}

		if len(peers) == 0 {
			fmt.Printf("No peers found. Make sure another device is running `dali %s`\n", openCmd)
			return nil
		}

		var peerIdx int
		if peerName != anyone && len(peers) == 1 {
			peerIdx = 0
		} else {
			numPeers := len(peers)
			if autoSend && numPeers == 1 {
				// Check if autosend to any 1 peer
				peerIdx = 0
			} else {
				// Let user select recipient
				fmt.Printf("\nFound %d peers:\n", numPeers)
				maxLength := maxPeerNameLength(peers)
				template := fmt.Sprintf("  [%%2d] %%%ds : %%s\n", maxLength)
				for i, peer := range peers {
					fmt.Printf(template, i+1, peer.Name, peer.Addr)
				}

				fmt.Println("\nEnter peer number to send to:")
				input := readInput()
				choice := number.ParseInt(input)
				if choice < 1 || choice > numPeers {
					return fmt.Errorf("invalid selection")
				}
				peerIdx = choice - 1
			}
		}

		peer := peers[peerIdx]
		peerName, peerAddr = peer.Name, peer.Addr
	}
	peer := Peer{Name: peerName, Addr: peerAddr}
	fmt.Printf("Sending %q to %s (%s)...\n", filePath, peerName, peerAddr)
	return sendFile(node, peer, filePath)
}

// Logs command handler
func cmdLogs(node *Node, options dict.StringMap) error {
	logs := node.Logs[:]
	slices.SortFunc(logs, func(e1, e2 Event) int {
		// Sort by descending timestamp
		return cmp.Compare(e2[0], e1[0])
	})
	fmt.Println("Logs:", len(logs))
	fromMaxLength := slices.Max(list.Map(logs, func(e Event) int {
		return len(e[5])
	}))
	toMaxLength := slices.Max(list.Map(logs, func(e Event) int {
		return len(e[6])
	}))
	template := fmt.Sprintf("%%s %%s %%s from=%%-%ds to=%%-%ds %%7s %%s\n", fromMaxLength, toMaxLength)
	for _, e := range logs {
		timestamp, event, result, path, size, sender, receiver := e.Tuple()
		event = str.Center(event, 8)
		result = str.Center(result, 7)
		size = computeFileSize(uint64(number.ParseInt(size)))
		fmt.Printf(template, timestamp, event, result, sender, receiver, size, path)
	}
	return nil
}
