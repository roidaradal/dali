package dali

import (
	"fmt"
	"net"
	"time"

	"github.com/roidaradal/fn/list"
)

const readDeadlineMs int = 200

type Peer struct {
	Name string
	Addr string
}

// DiscoverPeers broadcasts a query and collects peer responses
func DiscoverPeers(timeout time.Duration) ([]Peer, error) {
	// Create UDP socket for sending, address at 0.0.0.0:0 (port 0 = auto-select open port)
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to create UDP socket: %w", err)
	}
	defer conn.Close()

	// Send query to broadcast address, 255.255.255.255:<DISCOVERY_PORT>
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: DiscoveryPort,
	}
	query := newQueryMessage()
	_, err = conn.WriteToUDP(query.ToBytes(), broadcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send discovery query: %w", err)
	}

	// Collect responses
	var peers []Peer
	done := make(chan struct{}) // done channel

	// Goroutine for collecting responses
	readDuration := time.Duration(readDeadlineMs) * time.Millisecond
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-done:
				return // exit loop after timeout finishes
			default:
				conn.SetReadDeadline(time.Now().Add(readDuration))
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					continue
				}

				msg, err := parseDiscoveryMessage(buf[:n])
				if err != nil || msg.Type != announceType {
					continue // skip on error or non-Announcement messages
				}

				// mu.Lock() // from vibe code
				// TODO: Review this line: addr.IP.String() => what IP does it return?
				peerAddr := fmt.Sprintf("%s:%d", addr.IP.String(), msg.TransferPort)

				// Check if peer already exists
				exists := list.Any(peers, func(p Peer) bool {
					return p.Addr == peerAddr
				})
				if exists {
					continue
				}

				peers = append(peers, Peer{
					Name: msg.Name,
					Addr: peerAddr,
				})
				// mu.Unlock() // from vibe code
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
	// Create UDP socket for listening, address at 0.0.0.0:<DISCOVERY_PORT>
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: DiscoveryPort,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return fmt.Errorf("failed to bind discovery port: %w", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, peerAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg, err := parseDiscoveryMessage(buf[:n])
		if err != nil || msg.Type != queryType {
			continue // skip on error or non-Query messages
		}

		// Respond with our announcement
		name = compressName(name)
		announce := newAnnounceMessage(name, transferPort)
		conn.WriteToUDP(announce.ToBytes(), peerAddr)
	}
}
