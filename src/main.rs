mod discovery;
mod protocol;
mod transfer;

use clap::{Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser)]
#[command(name = "dali")]
#[command(about = "P2P local network file sharing tool")]
#[command(version)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Send a file to a peer on the network
    Send {
        /// File to send
        file: PathBuf,

        /// Target peer address (IP:PORT) - if not specified, discovers peers
        #[arg(short, long)]
        to: Option<String>,
    },
    /// Listen for incoming file transfers
    Receive {
        /// Output directory for received files
        #[arg(short, long, default_value = ".")]
        output: PathBuf,

        /// Port to listen on
        #[arg(short, long, default_value_t = protocol::TRANSFER_PORT)]
        port: u16,
    },
    /// Discover peers on the local network
    Discover {
        /// Discovery timeout in seconds
        #[arg(short, long, default_value_t = 3)]
        timeout: u64,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();

    match cli.command {
        Commands::Send { file, to } => {
            if !file.exists() {
                eprintln!("Error: File '{}' not found", file.display());
                std::process::exit(1);
            }

            let target = if let Some(addr) = to {
                addr
            } else {
                // Discover peers and let user select
                println!("Discovering peers...");
                let peers = discovery::discover_peers(3).await?;

                if peers.is_empty() {
                    eprintln!("No peers found. Make sure another device is running 'dali receive'.");
                    std::process::exit(1);
                }

                println!("\nFound peers:");
                for (i, peer) in peers.iter().enumerate() {
                    println!("  [{}] {} ({})", i + 1, peer.name, peer.addr);
                }

                println!("\nEnter peer number to send to:");
                let mut input = String::new();
                std::io::stdin().read_line(&mut input)?;
                let choice: usize = input.trim().parse().unwrap_or(0);

                if choice == 0 || choice > peers.len() {
                    eprintln!("Invalid selection");
                    std::process::exit(1);
                }

                peers[choice - 1].addr.to_string()
            };

            transfer::send_file(&target, &file).await?;
        }

        Commands::Receive { output, port } => {
            // Show connection info
            if let Ok(ip) = local_ip_address::local_ip() {
                println!("╔══════════════════════════════════════════╗");
                println!("║            dali - File Receiver          ║");
                println!("╠══════════════════════════════════════════╣");
                println!("║  Address: {:>28}  ║", format!("{}:{}", ip, port));
                println!("╚══════════════════════════════════════════╝");
            }

            // Start discovery listener in background
            let name = hostname::get()
                .map(|h| h.to_string_lossy().to_string())
                .unwrap_or_else(|_| "unknown".to_string());
            
            tokio::spawn(async move {
                let _ = discovery::run_discovery_listener(name, port).await;
            });

            // Start receiving files
            transfer::receive_files(port, &output).await?;
        }

        Commands::Discover { timeout } => {
            println!("Discovering peers on local network...\n");
            let peers = discovery::discover_peers(timeout).await?;

            if peers.is_empty() {
                println!("No peers found.");
            } else {
                println!("Found {} peer(s):", peers.len());
                for peer in peers {
                    println!("  • {} ({})", peer.name, peer.addr);
                }
            }
        }
    }

    Ok(())
}
