# dali 

A local network file sharing tool, written in Rust.

Dali (다리) means `bridge` in Korean, and also means `fast` in Filipino.

# TODO
* Set name 
* Add local logs for sending/receiving 
* Auto-send if only one peer discovered
* Add _1 for duplicate filenames (no override)
* Allow sending of files to computer name, not IPaddress
* Make GUI
* Create background process for listening 

# Test Scenarios
* 2 senders for 1 receiver, at the same time
* Overwritten old file with same filename