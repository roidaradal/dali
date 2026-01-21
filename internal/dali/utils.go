package dali

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/roidaradal/fn/dict"
	"github.com/roidaradal/fn/io"
	"github.com/roidaradal/fn/list"
	"github.com/roidaradal/fn/number"
	"github.com/schollz/progressbar/v3"
)

// Wildcard default for names and ip addresses
const anything string = "*"

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

// Create new IPv4 addres from ip string
func newIPv4(addr string) net.IP {
	parts := list.Map(strings.Split(addr, "."), number.ParseInt)
	a, b, c, d := byte(parts[0]), byte(parts[1]), byte(parts[2]), byte(parts[3])
	return net.IPv4(a, b, c, d)
}

// Get local IPv4 address
func getLocalIPv4Address(soloIP bool) (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", wrapErr("failed to get hostname", err)
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		return "", wrapErr("failed to get host IP addresses", err)
	}

	ips := make([]string, 0)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil && !ipv4.IsLoopback() {
			ips = append(ips, ipv4.String())
		}
	}

	if !soloIP {
		return strings.Join(ips, ", "), nil
	}

	numIPs := len(ips)
	switch numIPs {
	case 0:
		return "", nil
	case 1:
		return ips[0], nil
	default:
		// Display network choices
		name := getNetworkNames(ips)
		fmt.Println("\nSelect network:")
		for i, ip := range ips {
			fmt.Printf("  [%d] %s %s\n", i+1, ip, name[ip])
		}
		// Let user select network
		fmt.Println("\nEnter network number to use:")
		choice := number.ParseInt(readInput())
		if choice < 1 || choice > numIPs {
			return "", fmt.Errorf("invalid selection")
		}
		return ips[choice-1], nil
	}
}

// Get network interface names
func getNetworkNames(ips []string) dict.StringMap {
	name := make(dict.StringMap)
	for _, ip := range ips {
		name[ip] = "(?)"
	}
	interfaces, err := net.Interfaces()
	if err != nil {
		return name
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // skip if cannot get address
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet) // type-cast into IPNet
			if !ok {
				continue
			}
			ipv4 := ipnet.IP.To4()
			if ipv4 == nil {
				continue
			}
			key := ipv4.String()
			if _, ok := name[key]; ok {
				name[key] = fmt.Sprintf("(%s)", iface.Name)
				break
			}
		}
	}
	return name
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

var inputReader = bufio.NewReader(os.Stdin)

// Read input from stdin
func readInput() string {
	input, _ := inputReader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Convert the number of bytes to string (KB, MB, GB)
func computeFileSize(numBytes uint64) string {
	powers := []float64{
		math.Pow(1024, 3),
		math.Pow(1024, 2),
		math.Pow(1024, 1),
	}
	names := []string{"GB", "MB", "KB"}
	bytes := float64(numBytes)
	for i, denom := range powers {
		if bytes < denom {
			continue
		}
		value := bytes / denom
		return fmt.Sprintf("%.1f%s", value, names[i])
	}
	return fmt.Sprintf("%dB", numBytes)
}

// Find safe output file path (append _1, _2, ... if file already exists)
func getOutputPath(path string) string {
	if !io.PathExists(path) {
		return path
	}
	folder := filepath.Dir(path)
	filename := filepath.Base(path)
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	suffix := 1
	for {
		filename2 := fmt.Sprintf("%s_%d%s", name, suffix, ext)
		path2 := filepath.Join(folder, filename2)
		if !io.PathExists(path2) {
			return path2
		}
		suffix += 1
	}
}
