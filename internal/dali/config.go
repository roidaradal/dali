package dali

import (
	"fmt"
	"strings"

	"github.com/roidaradal/fn/io"
)

const (
	cfgPath        string = ".dali" // Full path: ~HOME/.dali
	defaultTimeout int    = 3       // Default timeout: 3s
	minTimeout     int    = 1       // Minimum timeout: 1s
	discoveryPort  int    = 45678   // UDP discovery port
	transferPort   uint16 = 45679   // TCP transfer port
)

// Type, FilePath, SenderName, SenderAddr, ReceiverName, ReceiverAddr
type Event [6]string

// User's configuration (name, timeout, logs)
type Config struct {
	Path    string `json:"-"`
	Name    string
	Timeout int
	Logs    []Event
}

// Representation of machine
type Node struct {
	*Config
	Addr string
}

// Create new Config
func newConfig(name string) *Config {
	return &Config{
		Name:    name,
		Timeout: defaultTimeout,
		Logs:    []Event{},
	}
}

// Save the config to file
func (c *Config) Save() error {
	// Save config file
	return io.SaveIndentedJSON(c, c.Path)
}

// Destructure event parts
func (e Event) Tuple() (eventType, filePath, senderName, senderAddr, receiverName, receiverAddr string) {
	eventType, filePath = e[0], e[1]
	senderName, senderAddr = e[2], e[3]
	receiverName, receiverAddr = e[4], e[5]
	return eventType, filePath, senderName, senderAddr, receiverName, receiverAddr
}

// String representationof Node
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
