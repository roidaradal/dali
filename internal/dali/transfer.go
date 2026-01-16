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

// Send file to specified address
func SendFile(addr string, filePath string) error {
	// Open file and get info
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := uint64(info.Size())
	fileName := filepath.Base(filePath)

	// Connect to peer
	fmt.Printf("Connecting to %s...\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to conect: %w", err)
	}
	defer conn.Close()

	// Send file offer
	offer := newOfferMessage(fileName, fileSize)
	_, err = conn.Write(offer.ToBytes())
	if err != nil {
		return fmt.Errorf("failed to send file offer: %w", err)
	}

	// Wait for response
	reader := bufio.NewReader(conn)
	responseLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	responseLine = strings.TrimSpace(responseLine)

	response, err := parseTransferMessage([]byte(responseLine))
	if err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	switch response.Type {
	case acceptType:
		fmt.Printf("Peer accepted. Sending %q...\n", fileName)
	case rejectType:
		fmt.Println("Peer rejected the file transfer.")
		return nil
	default:
		return fmt.Errorf("invalid response from peer: %s", response.Type)
	}

	// Send file data with progress bar
	bar := newProgressBar(fileSize, "Sending")
	buf := make([]byte, chunkSize)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}
		bar.Add(n)
	}

	fmt.Println("\n✓ File sent successfully!")
	return nil
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

	bar := newProgressBar(fileSize, "Receiving")

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
