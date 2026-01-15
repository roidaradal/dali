package dali

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/roidaradal/fn/io"
)

type Node struct {
	*Config
	Addr string
}

func (n Node) String() string {
	divider := strings.Repeat("=====", 5)
	out := []string{
		divider,
		fmt.Sprintf("Name: %s", n.Name),
		fmt.Sprintf("Addr: %s", n.Addr),
		fmt.Sprintf("Wait: %d", n.Timeout),
		divider,
	}
	return strings.Join(out, "\n")
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
		err = io.SaveIndentedJSON(newConfig(hostName), path)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize dali config: %w", err)
		}
	}

	// Load dali config file
	cfg, err := io.ReadJSON[Config](path)
	if err != nil {
		return nil, fmt.Errorf("failed to load dali config: %w", err)
	}
	cfg.Path = path

	addr, err := getLocalIPv4Address()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP addr: %w", err)
	}

	node := &Node{
		Config: cfg,
		Addr:   addr,
	}
	return node, nil
}
