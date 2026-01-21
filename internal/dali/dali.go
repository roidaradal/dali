// Package dali contains implementations of the dali tool
package dali

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/roidaradal/fn/clock"
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/lang"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/roidaradal/fn/str"
)

const currentVersion string = "0.1.4"

const (
	HelpCmd    string = "help"
	openCmd    string = "open"
	versionCmd string = "version"
	setCmd     string = "set"
	findCmd    string = "find"
	sendCmd    string = "send"
	updateCmd  string = "update"
	logsCmd    string = "logs"
	resetCmd   string = "reset"
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
	resetCmd:   cmdReset,
}

// List of commands, ordered for help
var commands = []string{setCmd, openCmd, sendCmd, findCmd, updateCmd, logsCmd, resetCmd, versionCmd, HelpCmd}

var cmdColor = map[string]func(string) string{
	HelpCmd:    str.Green,
	versionCmd: str.Yellow,
	setCmd:     str.Red,
	findCmd:    str.Cyan,
	openCmd:    str.Yellow,
	sendCmd:    str.Green,
	updateCmd:  str.Blue,
	logsCmd:    str.Violet,
	resetCmd:   str.Red,
}

var cmdSoloIP = map[string]bool{
	HelpCmd:    false,
	versionCmd: false,
	setCmd:     false,
	updateCmd:  false,
	logsCmd:    false,
	resetCmd:   false,
	findCmd:    true,
	openCmd:    true,
	sendCmd:    true,
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
	resetCmd:   "erase name, timeout, logs",
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
		{"file={FILE_PATH} wait", "wait for timeout to finish finding peers"},
	},
	findCmd: {
		{"", "look for all peers in local network"},
		{"name={NAME}", "look for peer {NAME} in local network"},
		{"ip={IP_ADDR}", "look for peer with specified IP address in local network"},
		{"wait", "wait for timeout to finish looking for peers"},
	},
	updateCmd: {
		{"", "update to latest version"},
		{"v=0.1.0", "update to specific version"},
		{"version=0.1.0", "update to specific version"},
	},
	logsCmd: {
		{"", "view all logs"},
		{"date={DATE}", "show logs for specified date"},
		{"action={ACTION}", "show 'send' or 'receive' logs"},
		{"from={NAME}", "show logs where sender is {NAME}"},
		{"to={NAME}", "show logs where receiver is {NAME}"},
		{"file={FILENAME}", "show logs where file path contains filename substring"},
	},
}

// Load user node
func LoadNode(command string) (*Node, error) {
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
	addr, err := getLocalIPv4Address(cmdSoloIP[command])
	if err != nil || addr == "" {
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
	// Options: name=NAME, ip=IPAddr, wait
	peerName, peerAddr := anything, anything
	endASAP := true
	for k, v := range options {
		switch k {
		case "name":
			// Make sure name has no spaces
			peerName = compressName(v)
		case "ip":
			peerAddr = v
		case "wait":
			endASAP = false
		}
	}

	fmt.Println(findingMessage(node))
	peers, err := discoverPeers(time.Duration(node.Timeout)*time.Second, Peer{Name: peerName, Addr: peerAddr}, endASAP)
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
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return wrapErr("failed to get absolute path of output dir", err)
	}
	fmt.Printf("Output folder: %s\n", absOutputDir)
	fmt.Printf("Listening for requests on local network at port %d...\n", listenPort)

	// Run discovery listener in the background
	go func() {
		runDiscoveryListener(node.Name, node.Addr, listenPort)
	}()

	err = receiveFiles(node, listenPort, outputDir, autoAccept, overwrite)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return nil
}

// Send command handler
func cmdSend(node *Node, options dict.StringMap) error {
	// Options: file=FILE_PATH, to=IPADDR:PORT, for=NAME, auto=1, wait
	filePath, peerAddr, peerName := "", "", anything
	autoSend := false
	endASAP := true
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
		case "wait":
			endASAP = false
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
		peers, err := discoverPeers(time.Duration(node.Timeout)*time.Second, Peer{Name: peerName, Addr: anything}, endASAP)
		if err != nil {
			return wrapErr("discovery failed", err)
		}

		if len(peers) == 0 {
			fmt.Printf("No peers found. Make sure another device is running `dali %s`\n", openCmd)
			return nil
		}

		var peerIdx int
		if peerName != anything && len(peers) == 1 {
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

				fmt.Printf("\nEnter peer number to send to: ")
				choice := number.ParseInt(readInput())
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
	// Options: date=DATE, action=ACTION, from=NAME, to=NAME, file=FILENAME_SUBSTRING
	filterDate, filterAction, filterFile := anything, anything, anything
	filterFrom, filterTo := anything, anything
	for k, v := range options {
		if v == "" {
			continue
		}
		switch k {
		case "date":
			filterDate = v
		case "action":
			filterAction = strings.ToLower(v)
		case "from":
			filterFrom = strings.ToLower(v)
		case "to":
			filterTo = strings.ToLower(v)
		case "file":
			filterFile = strings.ToLower(v)
		}
	}
	logs := list.Filter(node.Logs, func(e Event) bool {
		if filterDate != anything && clock.ExtractDate(e[evTimestamp]) != filterDate {
			return false
		}
		if filterAction != anything && e[evType] != filterAction {
			return false
		}
		if filterFrom != anything && strings.ToLower(e[evSender]) != filterFrom {
			return false
		}
		if filterTo != anything && strings.ToLower(e[evReceiver]) != filterTo {
			return false
		}
		if filterFile == anything {
			return true
		}
		pattern := regexp.MustCompile("(?i)" + filterFile)
		return pattern.MatchString(e[evPath])
	})
	numLogs := len(logs)
	fmt.Println("Logs:", numLogs)
	if numLogs == 0 {
		fmt.Println("No logs found")
		return nil
	}

	slices.SortFunc(logs, func(e1, e2 Event) int {
		// Sort by descending timestamp
		return cmp.Compare(e2[evTimestamp], e1[evTimestamp])
	})
	fromMaxLength := slices.Max(list.Map(logs, func(e Event) int {
		return len(e[evSender])
	}))
	toMaxLength := slices.Max(list.Map(logs, func(e Event) int {
		return len(e[evReceiver])
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

// Reset command handler
func cmdReset(node *Node, _ dict.StringMap) error {
	err := os.Remove(node.Config.Path)
	fmt.Println("dali reset", lang.Ternary(err == nil, "successful", "failed"))
	return err
}
