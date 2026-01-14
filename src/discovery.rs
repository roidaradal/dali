use crate::protocol::{DiscoveryMessage, Peer, DISCOVERY_PORT};
use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;
use tokio::net::UdpSocket;
use tokio::sync::RwLock;

/// Discover peers on the local network
pub async fn discover_peers(timeout_secs: u64) -> Result<Vec<Peer>, Box<dyn std::error::Error>> {
    let socket = UdpSocket::bind("0.0.0.0:0").await?;
    socket.set_broadcast(true)?;

    // Send query to broadcast address
    let query = DiscoveryMessage::Query;
    let broadcast_addr: SocketAddr = format!("255.255.255.255:{}", DISCOVERY_PORT).parse()?;
    socket.send_to(&query.to_bytes(), broadcast_addr).await?;

    let peers = Arc::new(RwLock::new(HashMap::new()));
    let peers_clone = peers.clone();

    // Listen for responses
    let listen_handle = tokio::spawn(async move {
        let mut buf = [0u8; 1024];
        loop {
            if let Ok((len, addr)) = socket.recv_from(&mut buf).await {
                if let Some(DiscoveryMessage::Announce { name, transfer_port }) =
                    DiscoveryMessage::from_bytes(&buf[..len])
                {
                    let peer_addr = SocketAddr::new(addr.ip(), transfer_port);
                    let mut peers = peers_clone.write().await;
                    peers.insert(addr.ip().to_string(), Peer { name, addr: peer_addr });
                }
            }
        }
    });

    // Wait for timeout
    tokio::time::sleep(Duration::from_secs(timeout_secs)).await;
    listen_handle.abort();

    let peers = peers.read().await;
    Ok(peers.values().cloned().collect())
}

/// Run discovery listener (announce our presence)
pub async fn run_discovery_listener(
    name: String,
    transfer_port: u16,
) -> Result<(), Box<dyn std::error::Error>> {
    let socket = UdpSocket::bind(format!("0.0.0.0:{}", DISCOVERY_PORT)).await?;
    socket.set_broadcast(true)?;

    let mut buf = [0u8; 1024];

    loop {
        let (len, addr) = socket.recv_from(&mut buf).await?;

        if let Some(msg) = DiscoveryMessage::from_bytes(&buf[..len]) {
            match msg {
                DiscoveryMessage::Query => {
                    // Respond with our announcement
                    let announce = DiscoveryMessage::Announce {
                        name: name.clone(),
                        transfer_port,
                    };
                    let _ = socket.send_to(&announce.to_bytes(), addr).await;
                }
                DiscoveryMessage::Announce { .. } => {
                    // Ignore announcements when listening
                }
            }
        }
    }
}
