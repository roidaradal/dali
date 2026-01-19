package dali

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/roidaradal/fn/clock"
)

// Chunk size for file transfer (64KB)
const chunkSize int = 64 * 1024

// Send file to specified address
func sendFile(node *Node, peer Peer, filePath string) error {
	// Open file and get info
	file, err := os.Open(filePath)
	if err != nil {
		return wrapErr("failed to open file", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return wrapErr("failed to get file info", err)
	}

	fileSize := uint64(info.Size())
	fileName := filepath.Base(filePath)
	size := fmt.Sprintf("%d", fileSize)

	// Connect to peer via TCP
	fmt.Printf("Connecting to %s...\n", peer.Addr)
	conn, err := net.Dial("tcp", peer.Addr)
	if err != nil {
		return wrapErr("failed to connect", err)
	}
	defer conn.Close()

	// Send file offer
	offer := newOfferMessage(node.Name, fileName, fileSize)
	_, err = conn.Write(offer.ToBytes())
	if err != nil {
		return wrapErr("failed to send file offer", err)
	}

	// Wait for response
	reader := bufio.NewReader(conn)
	responseLine, err := reader.ReadString('\n')
	if err != nil {
		return wrapErr("failed to read response", err)
	}
	responseLine = strings.TrimSpace(responseLine)

	response, err := parseMessage[TransferMessage]([]byte(responseLine))
	if err != nil {
		return wrapErr("invalid response", err)
	}

	// Create send event with empty result
	event := Event{clock.DateTimeNow(), "send", "", filePath, size, node.Name, peer.Name}

	// Check if responseType is 'accept'
	switch response.Type {
	case acceptType:
		fmt.Printf("Peer accepted. Sending %q...\n", fileName)
	case rejectType:
		addLog(node, event, rejectType)
		fmt.Println("Peer rejected the file transfer.")
		return nil
	default:
		addLog(node, event, "invalid")
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
			addLog(node, event, "fail")
			return wrapErr("failed to read file", err)
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			addLog(node, event, "fail")
			return wrapErr("failed to send data", err)
		}
		bar.Add(n)
	}

	addLog(node, event, "ok")
	fmt.Println("\n✓ File sent successfully!")
	return nil
}

// Listens for incoming file transfers
func receiveFiles(node *Node, port uint16, outputDir string, autoAccept bool) error {
	// Listen to port via TCP
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return wrapErr("failed to listen", err)
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
			if err := handleIncomingTransfer(node, c, outputDir, autoAccept); err != nil {
				fmt.Printf("Transfer error: %v\n", err)
			}
		}(conn)
	}
}

// Handle incoming file transfer
func handleIncomingTransfer(node *Node, conn net.Conn, outputDir string, autoAccept bool) error {
	reader := bufio.NewReader(conn)

	// Read file offer
	offerLine, err := reader.ReadString('\n')
	if err != nil {
		return wrapErr("failed to read offer", err)
	}
	offerLine = strings.TrimSpace(offerLine)

	offer, err := parseMessage[TransferMessage]([]byte(offerLine))
	if err != nil {
		return wrapErr("invalid offer", err)
	}

	if offer.Type != offerType {
		return fmt.Errorf("expected file offer, got %s", offer.Type)
	}

	fileName, fileSize := offer.Filename, offer.Size
	size := fmt.Sprintf("%d", fileSize)

	var msg *TransferMessage
	rejected := false
	if autoAccept {
		msg = newAcceptMessage()
	} else {
		// Prompt confirmation
		fmt.Printf("Incoming file %q (%d bytes) from %q. Accept? [Type 'N' to reject]: ", fileName, fileSize, offer.Sender)
		switch readInput() {
		case "N", "n":
			msg = newRejectMessage()
			rejected = true
		default:
			msg = newAcceptMessage()
		}
	}
	_, err = conn.Write(msg.ToBytes())
	if err != nil {
		return wrapErr("failed to send response", err)
	}
	outputPath := filepath.Join(outputDir, fileName)

	// Create receive event with empty result
	event := Event{clock.DateTimeNow(), "receive", "", outputPath, size, offer.Sender, node.Name}

	if rejected {
		addLog(node, event, rejectType)
		fmt.Println("Rejected file transfer.")
		return nil
	}

	// Receive file data
	fmt.Printf("Receiving %q (%d bytes)...\n", fileName, fileSize)
	file, err := os.Create(outputPath)
	if err != nil {
		addLog(node, event, "fail")
		return wrapErr("failed to create file", err)
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
			addLog(node, event, "fail")
			return wrapErr("failed to read data", err)
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			addLog(node, event, "fail")
			return wrapErr("failed to write file", err)
		}

		received += uint64(n)
		bar.Add(n)
	}

	addLog(node, event, "ok")
	fmt.Printf("\n✓ Saved to %q\n", outputPath)
	return nil
}

func addLog(node *Node, event Event, result string) {
	event[1] = result
	node.Config.AddLog(event)
	node.Config.Save()
}
