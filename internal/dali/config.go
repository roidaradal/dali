package dali

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/roidaradal/fn/io"
)

const cfgPath string = ".dali"

type Config struct {
	Name    string
	Timeout int
	Logs    []Event
}

// Type, FilePath, SenderName, SenderAddr, ReceiverName, ReceiverAddr
type Event [6]string

func (e Event) Tuple() (eventType, filePath, senderName, senderAddr, receiverName, receiverAddr string) {
	eventType, filePath = e[0], e[1]
	senderName, senderAddr = e[2], e[3]
	receiverName, receiverAddr = e[4], e[5]
	return eventType, filePath, senderName, senderAddr, receiverName, receiverAddr
}

type Node struct {
	Name string
	Addr string
}

// Load user node
func LoadNode() (*Node, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot load home dir: %w", err)
	}

	// Initialize dali config file, if it does not exist
	path := filepath.Join(homeDir, cfgPath)
	if !io.PathExists(path) {
		hostName, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("failed to get hostname: %w", err)
		}
		cfg := &Config{
			Name:    hostName,
			Timeout: 5, // default: 5 seconds
			Logs:    []Event{},
		}
		err = io.SaveIndentedJSON(cfg, path)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize dali config: %w", err)
		}
	}

	// Load dali config file
	cfg, err := io.ReadJSON[Config](path)
	if err != nil {
		return nil, fmt.Errorf("failed to load dali config: %w", err)
	}

	addr, err := getLocalIPv4Address()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP addr: %w", err)
	}

	node := &Node{
		Name: cfg.Name,
		Addr: addr,
	}
	return node, nil
}
