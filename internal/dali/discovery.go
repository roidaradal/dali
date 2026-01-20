package dali

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/roidaradal/fn/list"
)

const readDeadlineMs int = 100

type Peer struct {
	Name string
	Addr string
}

// DiscoverPeers broadcasts a query and collects peer responses
func discoverPeers(timeout time.Duration, filter Peer) ([]Peer, error) {
	// Create UDP socket for sending, address at 0.0.0.0:0 (port 0 = auto-select open port)
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, wrapErr("failed to create UDP socket", err)
	}
	defer conn.Close()

	// Send query to broadcast address, 255.255.255.255:<DISCOVERY_PORT>
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: discoveryPort,
	}
	query := newQueryMessage()
	_, err = conn.WriteToUDP(query.ToBytes(), broadcastAddr)
	if err != nil {
		return nil, wrapErr("failed to send discovery query", err)
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

				msg, err := parseMessage[DiscoveryMessage](buf[:n])
				if err != nil || msg.Type != announceType {
					continue // skip on error or non-Announcement messages
				}

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
			}
		}
	}()

	// Wait for timeout
	time.Sleep(timeout)
	close(done)

	if filter.Name != anything || filter.Addr != anything {
		targetName := strings.ToLower(filter.Name)
		targetAddr := fmt.Sprintf("%s:", filter.Addr)
		peers = list.Filter(peers, func(peer Peer) bool {
			ok1 := filter.Name == anything || strings.ToLower(peer.Name) == targetName
			ok2 := filter.Addr == anything || strings.HasPrefix(peer.Addr, targetAddr)
			return ok1 && ok2
		})
	}

	return peers, nil
}

// RunDiscoveryListener listens for discovery queries and responds with announcements
func runDiscoveryListener(name string, transferPort uint16) error {
	// Create UDP socket for listening, address at 0.0.0.0:<DISCOVERY_PORT>
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: discoveryPort,
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return wrapErr("failed to bind discovery port", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, peerAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg, err := parseMessage[DiscoveryMessage](buf[:n])
		if err != nil || msg.Type != queryType {
			continue // skip on error or non-Query messages
		}

		// Respond with our announcement
		name = compressName(name)
		announce := newAnnounceMessage(name, transferPort)
		conn.WriteToUDP(announce.ToBytes(), peerAddr)
	}
}
