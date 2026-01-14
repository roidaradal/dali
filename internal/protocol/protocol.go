package protocol

import "encoding/json"

// UDP discovery port
const DiscoveryPort = 45678

// TCP transfer port
const TransferPort uint16 = 45679

// Chunk size for file transfer (64KB)
const ChunkSize = 64 * 1024

// DiscoveryMessage represents messages sent during peer discovery
type DiscoveryMessage struct {
	Type         string `json:"type"`                    // "query" or "announce"
	Name         string `json:"name,omitempty"`          // peer name (for announce)
	TransferPort uint16 `json:"transfer_port,omitempty"` // transfer port (for announce)
}

// Peer represents a discovered peer on the network
type Peer struct {
	Name string
	Addr string
}

// TransferMessage represents messages during file transfer
type TransferMessage struct {
	Type     string `json:"type"`               // "file_offer", "accept", "reject", "complete"
	Filename string `json:"filename,omitempty"` // filename (for file_offer)
	Size     uint64 `json:"size,omitempty"`     // file size (for file_offer)
}

// NewQueryMessage creates a discovery query message
func NewQueryMessage() *DiscoveryMessage {
	return &DiscoveryMessage{Type: "Query"}
}

// NewAnnounceMessage creates a discovery announce message
func NewAnnounceMessage(name string, transferPort uint16) *DiscoveryMessage {
	return &DiscoveryMessage{
		Type:         "Announce",
		Name:         name,
		TransferPort: transferPort,
	}
}

// NewFileOfferMessage creates a file offer message
func NewFileOfferMessage(filename string, size uint64) *TransferMessage {
	return &TransferMessage{
		Type:     "FileOffer",
		Filename: filename,
		Size:     size,
	}
}

// NewAcceptMessage creates an accept message
func NewAcceptMessage() *TransferMessage {
	return &TransferMessage{Type: "Accept"}
}

// NewRejectMessage creates a reject message
func NewRejectMessage() *TransferMessage {
	return &TransferMessage{Type: "Reject"}
}

// ToBytes serializes the discovery message to JSON bytes
func (m *DiscoveryMessage) ToBytes() []byte {
	data, _ := json.Marshal(m)
	return data
}

// FromBytes deserializes a discovery message from JSON bytes
func ParseDiscoveryMessage(data []byte) (*DiscoveryMessage, error) {
	var msg DiscoveryMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ToBytes serializes the transfer message to JSON bytes with newline
func (m *TransferMessage) ToBytes() []byte {
	data, _ := json.Marshal(m)
	data = append(data, '\n')
	return data
}

// ParseTransferMessage deserializes a transfer message from JSON bytes
func ParseTransferMessage(data []byte) (*TransferMessage, error) {
	var msg TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
