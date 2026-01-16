# dali 

A local network peer-to-peer file sharing tool, written in Go.

Dali (다리) means `bridge` in Korean, and also means `fast` in Filipino.

## Installation

`go install github.com/roidaradal/dali@latest`

OR

Download `dali.exe` from the [releases](https://github.com/roidaradal/dali/releases) page. Add the folder where you saved `dali.exe` to your system PATH.

## Usage

## TODO
* Help and Version command
* Add local logs for sending/receiving 
* View logs
* find with name, IPaddress options
* Auto-send if only one peer discovered (autosend flag)
* Add _1 for duplicate filenames (no override)
* Allow sending of files to name, not IPaddress
* Make GUI
* Create background process for listening 
* Allow viewing of shared folder in network
* Allow fetching of file in shared folder

## Test Scenarios
* 2 senders for 1 receiver, at the same time
* Overwritten old file with same filename
* 1 shared folder, multiple downloaders at same time