package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/roidaradal/dali/internal/discovery"
	"github.com/roidaradal/dali/internal/protocol"
	"github.com/roidaradal/dali/internal/transfer"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dali",
		Short: "P2P local network file sharing tool",
	}

	// Send command
	var sendTo string
	sendCmd := &cobra.Command{
		Use:   "send <file>",
		Short: "Send a file to a peer on the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check if file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return fmt.Errorf("file '%s' not found", filePath)
			}

			var target string
			if sendTo != "" {
				target = sendTo
			} else {
				// Discover peers and let user select
				fmt.Println("Discovering peers...")
				peers, err := discovery.DiscoverPeers(3 * time.Second)
				if err != nil {
					return fmt.Errorf("discovery failed: %w", err)
				}

				if len(peers) == 0 {
					return fmt.Errorf("no peers found. Make sure another device is running 'dali receive'")
				}

				fmt.Println("\nFound peers:")
				for i, peer := range peers {
					fmt.Printf("  [%d] %s (%s)\n", i+1, peer.Name, peer.Addr)
				}

				fmt.Println("\nEnter peer number to send to:")
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				choice, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil || choice < 1 || choice > len(peers) {
					return fmt.Errorf("invalid selection")
				}

				target = peers[choice-1].Addr
			}

			return transfer.SendFile(target, filePath)
		},
	}
	sendCmd.Flags().StringVarP(&sendTo, "to", "t", "", "Target peer address (IP:PORT)")
	rootCmd.AddCommand(sendCmd)

	// Receive command
	var outputDir string
	var receivePort uint16
	receiveCmd := &cobra.Command{
		Use:   "receive",
		Short: "Listen for incoming file transfers",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show connection info
			if ip := getLocalIP(); ip != "" {
				fmt.Println("╔══════════════════════════════════════════╗")
				fmt.Println("║            dali - File Receiver          ║")
				fmt.Println("╠══════════════════════════════════════════╣")
				fmt.Printf("║  Address: %28s  ║\n", fmt.Sprintf("%s:%d", ip, receivePort))
				fmt.Println("╚══════════════════════════════════════════╝")
			}

			// Start discovery listener in background
			hostname, _ := os.Hostname()
			if hostname == "" {
				hostname = "unknown"
			}

			go func() {
				discovery.RunDiscoveryListener(hostname, receivePort)
			}()

			// Start receiving files
			return transfer.ReceiveFiles(receivePort, outputDir)
		},
	}
	receiveCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for received files")
	receiveCmd.Flags().Uint16VarP(&receivePort, "port", "p", protocol.TransferPort, "Port to listen on")
	rootCmd.AddCommand(receiveCmd)

	// Discover command
	var discoverTimeout uint64
	discoverCmd := &cobra.Command{
		Use:   "discover",
		Short: "Discover peers on the local network",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Discovering peers on local network...")
			peers, err := discovery.DiscoverPeers(time.Duration(discoverTimeout) * time.Second)
			if err != nil {
				return err
			}

			if len(peers) == 0 {
				fmt.Println("No peers found.")
			} else {
				fmt.Printf("Found %d peer(s):\n", len(peers))
				for _, peer := range peers {
					fmt.Printf("  • %s (%s)\n", peer.Name, peer.Addr)
				}
			}
			return nil
		},
	}
	discoverCmd.Flags().Uint64VarP(&discoverTimeout, "timeout", "t", 3, "Discovery timeout in seconds")
	rootCmd.AddCommand(discoverCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// getLocalIP returns the local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
