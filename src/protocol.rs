use serde::{Deserialize, Serialize};
use std::net::SocketAddr;

/// UDP discovery port
pub const DISCOVERY_PORT: u16 = 45678;

/// TCP transfer port
pub const TRANSFER_PORT: u16 = 45679;

/// Discovery message types
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum DiscoveryMessage {
    /// Announce presence on network
    Announce { name: String, transfer_port: u16 },
    /// Query for other peers
    Query,
}

/// Peer information
#[derive(Debug, Clone)]
pub struct Peer {
    pub name: String,
    pub addr: SocketAddr,
}

/// Transfer protocol messages
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum TransferMessage {
    /// Request to send a file
    FileOffer {
        filename: String,
        size: u64,
    },
    /// Accept file transfer
    Accept,
    /// Reject file transfer
    Reject,
    /// Transfer complete
    Complete,
}

impl DiscoveryMessage {
    pub fn to_bytes(&self) -> Vec<u8> {
        serde_json::to_vec(self).unwrap()
    }

    pub fn from_bytes(bytes: &[u8]) -> Option<Self> {
        serde_json::from_slice(bytes).ok()
    }
}

impl TransferMessage {
    pub fn to_bytes(&self) -> Vec<u8> {
        let mut json = serde_json::to_vec(self).unwrap();
        json.push(b'\n');
        json
    }

    pub fn from_bytes(bytes: &[u8]) -> Option<Self> {
        serde_json::from_slice(bytes).ok()
    }
}
