package dali

import (
	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/number"
)

// Full path: ~HOME/.dali
const cfgPath string = ".dali"

// Default timeout: 5 seconds
const defaultTimeout int = 5

type Config struct {
	Path    string `json:"-"`
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

func newConfig(name string) *Config {
	return &Config{
		Name:    name,
		Timeout: defaultTimeout,
		Logs:    []Event{},
	}
}

func (c *Config) Update(options dict.StringMap) error {
	for k, v := range options {
		switch k {
		case "name":
			c.Name = compressName(v)
		case "timeout", "wait":
			c.Timeout = max(defaultTimeout, number.ParseInt(v))
		}
	}
	return io.SaveIndentedJSON(c, c.Path)
}
