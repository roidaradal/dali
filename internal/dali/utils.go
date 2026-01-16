package dali

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/roidaradal/fn/list"
	"github.com/schollz/progressbar/v3"
)

// Wildcard default for names and ip addresses
const anyone string = "*"

// Remove spaces from name
func compressName(name string) string {
	return strings.Join(strings.Fields(name), "")
}

// Get max peer name length from list of Peers
func maxPeerNameLength(peers []Peer) int {
	return slices.Max(list.Map(peers, func(p Peer) int {
		return len(p.Name)
	}))
}

// Wrap error with prefix message
func wrapErr(message string, err error) error {
	return fmt.Errorf("%s: %w", message, err)
}

// Get local IPv4 address
func getLocalIPv4Address() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil && !ipv4.IsLoopback() {
			return ipv4.String(), nil
		}
	}

	return "", nil
}

// Create new progress bar
func newProgressBar(fileSize uint64, title string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		int64(fileSize),
		progressbar.OptionSetDescription(title),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

// Create finding peers message
func findingMessage(node *Node) string {
	return fmt.Sprintf("Finding peers on local network for %ds...\n", node.Timeout)
}
