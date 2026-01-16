package dali

import "encoding/json"

const (
	queryType    string = "query"
	announceType string = "announce"
	offerType    string = "offer"
	acceptType   string = "accept"
	rejectType   string = "reject"
)

type DiscoveryMessage struct {
	Type         string // query, announce
	Name         string // peer name (for announce)
	TransferPort uint16 // transfer port (for announce)
}

type TransferMessage struct {
	Type     string // offer, accept, reject, complete
	Filename string // file name (for offer)
	Size     uint64 // file size (for offer)
}

// Create new query DiscoveryMessage
func newQueryMessage() *DiscoveryMessage {
	return &DiscoveryMessage{
		Type: queryType,
	}
}

// Create new announce DiscoveryMessage
func newAnnounceMessage(name string, transferPort uint16) *DiscoveryMessage {
	return &DiscoveryMessage{
		Type:         announceType,
		Name:         name,
		TransferPort: transferPort,
	}
}

// Create new offer TransferMessage
func newOfferMessage(filename string, size uint64) *TransferMessage {
	return &TransferMessage{
		Type:     offerType,
		Filename: filename,
		Size:     size,
	}
}

// Create new accept TransferMessage
func newAcceptMessage() *TransferMessage {
	return &TransferMessage{Type: acceptType}
}

// Deserialize DiscoveryMessage from JSON bytes
func parseDiscoveryMessage(data []byte) (*DiscoveryMessage, error) {
	var msg DiscoveryMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Deserialize TransferMessage from JSON bytes
func parseTransferMessage(data []byte) (*TransferMessage, error) {
	var msg TransferMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// Serialize DiscoveryMessage to JSON bytes
func (m *DiscoveryMessage) ToBytes() []byte {
	data, _ := json.Marshal(m)
	return data
}

// Serialize TransferMessage to JSON bytes with newline
func (m *TransferMessage) ToBytes() []byte {
	data, _ := json.Marshal(m)
	data = append(data, '\n')
	return data
}
