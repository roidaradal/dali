package dali

import (
	"fmt"
	"strings"

	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/str"
)

const (
	cfgPath        string = ".dali" // Full path: ~HOME/.dali
	defaultTimeout int    = 3       // Default timeout: 3s
	minTimeout     int    = 1       // Minimum timeout: 1s
	discoveryPort  int    = 45678   // UDP discovery port
	transferPort   uint16 = 45679   // TCP transfer port
)

// Timestamp, Type, Result, FilePath, FileSize, SenderName, ReceiverName
type Event [7]string

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

// Add event log to config
func (c *Config) AddLog(event Event) {
	c.Logs = append(c.Logs, event)
}

// Save the config to file
func (c *Config) Save() error {
	// Save config file
	return io.SaveJSON(c, c.Path)
}

// Destructure event parts
func (e Event) Tuple() (timestamp, eventType, result, filePath, fileSize, senderName, receiverName string) {
	timestamp, eventType, result = e[0], e[1], e[2]
	filePath, fileSize = e[3], e[4]
	senderName, receiverName = e[5], e[6]
	return timestamp, eventType, result, filePath, fileSize, senderName, receiverName
}

// String representationof Node
func (n Node) String() string {
	divider := strings.Repeat("=====", 5)
	out := []string{
		divider,
		fmt.Sprintf("Name: %s", str.Green(n.Name)),
		fmt.Sprintf("Addr: %s", str.Yellow(n.Addr)),
		fmt.Sprintf("Wait: %s", str.Red(str.Int(n.Timeout))),
		divider,
	}
	return strings.Join(out, "\n")
}
