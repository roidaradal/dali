# dali 

A local network file sharing tool, written in Go.

Dali (다리) means `bridge` in Korean, and also means `fast` in Filipino.

## Installation

Download `dali.exe` from the [releases](https://github.com/roidaradal/dali/releases) page. Add the folder where you saved `dali.exe` to your system PATH.

## Usage

### Receive files

Start the receiver on a machine to accept incoming file transfers:

```bash
dali receive                     # Listen on default port (45679)
dali receive --port 8080         # Listen on custom port
dali receive --output ./downloads # Save files to specific folder
```

### Send files

Send a file to another machine running `dali receive`:

```bash
dali send myfile.txt             # Discover peers and select one
dali send myfile.txt --to 192.168.1.100:45679  # Send to specific address
```

### Discover peers

Find other machines running `dali receive` on the local network:

```bash
dali discover                    # 3 second timeout (default)
dali discover --timeout 10       # 10 second timeout
```

## TODO
* Set name  
* Allow setting of output folder
* Allow viewing of shared folder in network
* Allow fetching of file in shared folder
* Add local logs for sending/receiving 
* Auto-send if only one peer discovered
* Add _1 for duplicate filenames (no override)
* Allow sending of files to computer name, not IPaddress
* Make GUI
* Create background process for listening 

## Test Scenarios
* 2 senders for 1 receiver, at the same time
* Overwritten old file with same filename
* 1 shared folder, multiple downloaders at same time