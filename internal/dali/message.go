package dali

import "encoding/json"

const (
	queryType    string = "query"
	announceType string = "announce"
)

type DiscoveryMessage struct {
	Type         string // query, announce
	Name         string // peer name (for announce)
	TransferPort uint16 // transfer port (for announce)
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

// Deserialize DiscoveryMessage from JSON bytes
func parseDiscoveryMessage(data []byte) (*DiscoveryMessage, error) {
	var msg DiscoveryMessage
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
