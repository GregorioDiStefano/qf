# QF

A simple data transfer tool which uses AWS/GCP object storage as a backend. No need to create elaborate SSH tunnels to move files between machines!  Simply pipe data/file into `qf` and request the ID on the other machine(s).

- [x] Transparent encryption/decryption (PGP & SSL)
- [x] Transparent compression (GZIP)
- [X] Streaming transfer (start downloading before upload completed)
- [X] No SSH tunnels, firewall traversal

### Basic usage:

Upload a file: 

`cat file | qf ` (produces a unique ID)

To download, just pass the unique as a single argument and either print to stdout or redirect to a file:


`qf someID > file`

### Build

In order to use this tool with your own AWS or GCP account, you need to build the tool using the the `build.sh` script.

The script will build the `qf` binary with your AWS/GCP credentials embedded, making the tool fully independant."

### Demo

Moving a file from a laptop to a remote VM instance 


[![asciicast](https://asciinema.org/a/Ro940E0Lbq8uXKDD8JCLyZuxP.png)](https://asciinema.org/a/Ro940E0Lbq8uXKDD8JCLyZuxP)
