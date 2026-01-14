use crate::protocol::TransferMessage;
use indicatif::{ProgressBar, ProgressStyle};
use std::path::Path;
use tokio::fs::File;
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};

const CHUNK_SIZE: usize = 64 * 1024; // 64KB chunks

/// Send a file to a peer
pub async fn send_file(
    addr: &str,
    file_path: &Path,
) -> Result<(), Box<dyn std::error::Error>> {
    let file = File::open(file_path).await?;
    let metadata = file.metadata().await?;
    let file_size = metadata.len();
    let filename = file_path
        .file_name()
        .and_then(|n| n.to_str())
        .unwrap_or("unknown")
        .to_string();

    println!("Connecting to {}...", addr);
    let mut stream = TcpStream::connect(addr).await?;

    // Send file offer
    let offer = TransferMessage::FileOffer {
        filename: filename.clone(),
        size: file_size,
    };
    stream.write_all(&offer.to_bytes()).await?;

    // Wait for response
    let mut reader = BufReader::new(&mut stream);
    let mut response_line = String::new();
    reader.read_line(&mut response_line).await?;
    
    let response = TransferMessage::from_bytes(response_line.trim().as_bytes());

    match response {
        Some(TransferMessage::Accept) => {
            println!("Peer accepted. Sending '{}'...", filename);
        }
        Some(TransferMessage::Reject) => {
            println!("Peer rejected the file transfer.");
            return Ok(());
        }
        _ => {
            println!("Invalid response from peer.");
            return Ok(());
        }
    }

    // Send file data
    let pb = ProgressBar::new(file_size);
    pb.set_style(
        ProgressStyle::default_bar()
            .template("{spinner:.green} [{bar:40.cyan/blue}] {bytes}/{total_bytes} ({eta})")?
            .progress_chars("█▓░"),
    );

    let mut file = File::open(file_path).await?;
    let mut buf = vec![0u8; CHUNK_SIZE];
    let stream = reader.into_inner();

    loop {
        let n = file.read(&mut buf).await?;
        if n == 0 {
            break;
        }
        stream.write_all(&buf[..n]).await?;
        pb.inc(n as u64);
    }

    pb.finish_with_message("Transfer complete!");
    println!("✓ File sent successfully!");

    Ok(())
}

/// Listen for incoming file transfers
pub async fn receive_files(
    port: u16,
    output_dir: &Path,
) -> Result<(), Box<dyn std::error::Error>> {
    let listener = TcpListener::bind(format!("0.0.0.0:{}", port)).await?;
    println!("Listening for incoming files on port {}...", port);

    loop {
        let (mut stream, addr) = listener.accept().await?;
        println!("\nIncoming connection from {}", addr);

        let output_dir = output_dir.to_path_buf();

        tokio::spawn(async move {
            if let Err(e) = handle_incoming_transfer(&mut stream, &output_dir).await {
                eprintln!("Transfer error: {}", e);
            }
        });
    }
}

async fn handle_incoming_transfer(
    stream: &mut TcpStream,
    output_dir: &Path,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let mut reader = BufReader::new(stream);
    let mut offer_line = String::new();
    reader.read_line(&mut offer_line).await?;

    let offer = TransferMessage::from_bytes(offer_line.trim().as_bytes());

    let (filename, file_size) = match offer {
        Some(TransferMessage::FileOffer { filename, size }) => (filename, size),
        _ => {
            return Err("Invalid file offer".into());
        }
    };

    println!("Receiving '{}' ({} bytes)", filename, file_size);

    // Auto-accept for now
    let stream = reader.into_inner();
    let accept = TransferMessage::Accept;
    stream.write_all(&accept.to_bytes()).await?;

    // Receive file data
    let output_path = output_dir.join(&filename);
    let mut file = File::create(&output_path).await?;

    let pb = ProgressBar::new(file_size);
    pb.set_style(
        ProgressStyle::default_bar()
            .template("{spinner:.green} [{bar:40.cyan/blue}] {bytes}/{total_bytes} ({eta})")?
            .progress_chars("█▓░"),
    );

    let mut buf = vec![0u8; CHUNK_SIZE];
    let mut received: u64 = 0;

    while received < file_size {
        let to_read = std::cmp::min(CHUNK_SIZE as u64, file_size - received) as usize;
        let n = stream.read(&mut buf[..to_read]).await?;
        if n == 0 {
            break;
        }
        file.write_all(&buf[..n]).await?;
        received += n as u64;
        pb.inc(n as u64);
    }

    pb.finish_with_message("Complete!");
    println!("✓ Saved to '{}'", output_path.display());

    Ok(())
}
