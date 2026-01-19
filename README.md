# dali 

A local network peer-to-peer file transfer tool, written in Go.

Dali (다리) means `bridge` in Korean, and also means `fast` in Filipino.

## Installation

`go install github.com/roidaradal/dali@latest`

OR

Download `dali.exe` from the [releases](https://github.com/roidaradal/dali/releases) page. Add the folder where you saved `dali.exe` to your system PATH.

## Usage

### Set config 

Set your name and waiting time (in seconds) for finding peers. This data is saved in `~/.dali`

```bash
dali set name={NAME}                # Set your name (no spaces)
dali set wait={TIMEOUT_SECS}        # Set waiting time (in seconds) for finding peers
dali set timeout={TIMEOUT_SECS}     # Set waiting time (in seconds) for finding peers
```

### Receive files 

Open the machine to receive files and be discovered in peer queries:

```bash 
dali open                       # listen on default port (45679)
dali open port={PORT}           # listen on custom port
dali open out={OUT_DIR}         # listen and set output folder
dali open output={OUT_DIR}      # listen and set output folder
dali open accept=auto           # auto-accepts incoming file transfers
dali open overwrite             # overwrite old file path if it exists
```

### Find peers 

Find machines running `dali open` on the local network:

```bash 
dali find               # Look for peers in the local network for (timeout) seconds
dali find name={NAME}   # Look for peer named {NAME} in local network
dali find ip={IP_ADDR}  # Look for peer with specified IP address in local network
```

### Send file 

Send a file to another machine runing `dali open`:

```bash
dali send file={FILE_PATH}                  # Finds peers and select one to send file to
dali send file={FILE_PATH} for={NAME}       # Find peer named {NAME} and send file
dali send file={FILE_PATH} to={IPADDR:PORT} # Send file to specific address in local network
dali send file={FILE_PATH} auto=1           # Send file automatically if only 1 peer found
```

### Update 

Update dali to latest (or specific) version:

```bash
dali update                 # Update dali to latest version
dali update v=0.1.0         # Update dali to specific version
dali update version=0.1.0   # Update dali to specific version
```

### Other commands 

```bash
dali help       # Display help message
dali version    # Display current version
```

## TODO
* Make GUI
* Create background process for listening 
* Allow viewing of shared folder in network
* Allow fetching of file in shared folder
* Fix: machine has multiple IPv4 addresses (WiFi, LAN)

## Test Scenarios
* 2 senders for 1 receiver, at the same time
* Overwritten old file with same filename
* 1 shared folder, multiple downloaders at same time