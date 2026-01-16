package dali

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// Chunk size for file transfer (64KB)
const chunkSize int = 64 * 1024

var progressTheme = progressbar.Theme{
	Saucer:        "█",
	SaucerHead:    "▓",
	SaucerPadding: "░",
	BarStart:      "[",
	BarEnd:        "]",
}

// Listens for incoming file transfers
func ReceiveFiles(port uint16, outputDir string) error {
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		fmt.Printf("\nIncoming connection from %s...\n", conn.RemoteAddr())

		go func(c net.Conn) {
			defer c.Close()
			if err := handleIncomingTransfer(c, outputDir); err != nil {
				fmt.Printf("Transfer error: %v\n", err)
			}
		}(conn)
	}
}

// Handle incoming file transfer
func handleIncomingTransfer(conn net.Conn, outputDir string) error {
	reader := bufio.NewReader(conn)

	// Read file offer
	offerLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read offer: %w", err)
	}
	offerLine = strings.TrimSpace(offerLine)

	offer, err := parseTransferMessage([]byte(offerLine))
	if err != nil {
		return fmt.Errorf("invalid offer: %w", err)
	}

	if offer.Type != offerType {
		return fmt.Errorf("expected file offer, got %s", offer.Type)
	}

	fileName, fileSize := offer.Filename, offer.Size
	fmt.Printf("Receiving %q (%d bytes)...\n", fileName, fileSize)

	// Auto-accept
	_, err = conn.Write(newAcceptMessage().ToBytes())
	if err != nil {
		return fmt.Errorf("failed to send accept: %w", err)
	}

	// Receive file data
	outputPath := filepath.Join(outputDir, fileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	bar := progressbar.NewOptions64(
		int64(fileSize),
		progressbar.OptionSetDescription("Receiving"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressTheme),
	)

	buf := make([]byte, chunkSize)
	var received uint64

	for received < fileSize {
		toRead := chunkSize
		if remaining := fileSize - received; remaining < uint64(toRead) {
			toRead = int(remaining)
		}

		n, err := reader.Read(buf[:toRead])
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read data: %w", err)
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		received += uint64(n)
		bar.Add(n)
	}

	fmt.Printf("\n✓ Saved to %q\n", outputPath)
	return nil
}
