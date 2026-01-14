package transfer

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"

	"github.com/roidaradal/dali/internal/protocol"
)

// SendFile sends a file to the specified address
func SendFile(addr string, filePath string) error {
	// Open and get file info
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := uint64(fileInfo.Size())
	filename := filepath.Base(filePath)

	// Connect to peer
	fmt.Printf("Connecting to %s...\n", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Send file offer
	offer := protocol.NewFileOfferMessage(filename, fileSize)
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

	response, err := protocol.ParseTransferMessage([]byte(strings.TrimSpace(responseLine)))
	if err != nil {
		return fmt.Errorf("invalid response: %w", err)
	}

	switch response.Type {
	case "Accept":
		fmt.Printf("Peer accepted. Sending '%s'...\n", filename)
	case "Reject":
		fmt.Println("Peer rejected the file transfer.")
		return nil
	default:
		fmt.Println("Invalid response from peer.")
		return nil
	}

	// Send file data with progress bar
	bar := progressbar.NewOptions64(
		int64(fileSize),
		progressbar.OptionSetDescription("Sending"),
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

	buf := make([]byte, protocol.ChunkSize)
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

// ReceiveFiles listens for incoming file transfers
func ReceiveFiles(port uint16, outputDir string) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer listener.Close()

	fmt.Printf("Listening for incoming files on port %d...\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		fmt.Printf("\nIncoming connection from %s\n", conn.RemoteAddr())

		go func(c net.Conn) {
			defer c.Close()
			if err := handleIncomingTransfer(c, outputDir); err != nil {
				fmt.Printf("Transfer error: %v\n", err)
			}
		}(conn)
	}
}

func handleIncomingTransfer(conn net.Conn, outputDir string) error {
	reader := bufio.NewReader(conn)

	// Read file offer
	offerLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read offer: %w", err)
	}

	offer, err := protocol.ParseTransferMessage([]byte(strings.TrimSpace(offerLine)))
	if err != nil {
		return fmt.Errorf("invalid offer: %w", err)
	}

	if offer.Type != "FileOffer" {
		return fmt.Errorf("expected file offer, got %s", offer.Type)
	}

	filename := offer.Filename
	fileSize := offer.Size

	fmt.Printf("Receiving '%s' (%d bytes)\n", filename, fileSize)

	// Auto-accept for now
	accept := protocol.NewAcceptMessage()
	_, err = conn.Write(accept.ToBytes())
	if err != nil {
		return fmt.Errorf("failed to send accept: %w", err)
	}

	// Receive file data
	outputPath := filepath.Join(outputDir, filename)
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
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	buf := make([]byte, protocol.ChunkSize)
	var received uint64

	for received < fileSize {
		toRead := protocol.ChunkSize
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

	fmt.Printf("\n✓ Saved to '%s'\n", outputPath)
	return nil
}
