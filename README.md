# dali 

A local network peer-to-peer file sharing tool, written in Go.

Dali (다리) means `bridge` in Korean, and also means `fast` in Filipino.

## Installation

`go install github.com/roidaradal/dali@latest`

OR

Download `dali.exe` from the [releases](https://github.com/roidaradal/dali/releases) page. Add the folder where you saved `dali.exe` to your system PATH.

## Usage

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