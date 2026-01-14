package discovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/roidaradal/dali/internal/protocol"
)

// DiscoverPeers broadcasts a query and collects peer responses
func DiscoverPeers(timeout time.Duration) ([]protocol.Peer, error) {
	// Create UDP socket for sending
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}
	defer conn.Close()

	// Send query to broadcast address
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: protocol.DiscoveryPort,
	}

	query := protocol.NewQueryMessage()
	_, err = conn.WriteToUDP(query.ToBytes(), broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send discovery query: %w", err)
	}

	// Collect responses
	var peers []protocol.Peer
	var mu sync.Mutex
	done := make(chan struct{})

	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-done:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					continue
				}

				msg, err := protocol.ParseDiscoveryMessage(buf[:n])
				if err != nil {
					continue
				}

				if msg.Type == "Announce" {
					peerAddr := fmt.Sprintf("%s:%d", addr.IP.String(), msg.TransferPort)
					mu.Lock()
					// Check if peer already exists
					exists := false
					for _, p := range peers {
						if p.Addr == peerAddr {
							exists = true
							break
						}
					}
					if !exists {
						peers = append(peers, protocol.Peer{
							Name: msg.Name,
							Addr: peerAddr,
						})
					}
					mu.Unlock()
				}
			}
		}
	}()

	// Wait for timeout
	time.Sleep(timeout)
	close(done)

	return peers, nil
}

// RunDiscoveryListener listens for discovery queries and responds with announcements
func RunDiscoveryListener(name string, transferPort uint16) error {
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: protocol.DiscoveryPort,
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("failed to bind discovery port: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg, err := protocol.ParseDiscoveryMessage(buf[:n])
		if err != nil {
			continue
		}

		if msg.Type == "Query" {
			// Respond with our announcement
			announce := protocol.NewAnnounceMessage(name, transferPort)
			conn.WriteToUDP(announce.ToBytes(), remoteAddr)
		}
	}
}
