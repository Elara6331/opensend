# OpenSend
## Send files between systems quickly and securely

### Usage

#### Receiver
- Use `opensend -r` to start the receiver

#### Sender
- Use `opensend -s -t <type> -d <data>`
- `type` can either be `url` or `file`
- If `type` is `url`, `data` should be a URL
- If `type` is `file`, `data` should be a file path
- Example: `opensend -s -t url -d "https://google.com"`
- Example: `opensend -s -t file -d ~/file.txt`

### Building
- This project uses go modules, so building is easy
- First, go 1.14+ must be installed (use buster-backports on debian)
- Then, run `make` inside the project's directory.
- This will get the dependencies and compile all the files.

### Installing
To install, simply follow the building instructions and then run
- Linux: `sudo make install`
- macOS: `sudo make install-macos`

### Using on iOS
Opensend can run on iOS using the [iSH app](https://apps.apple.com/us/app/ish-shell/id1436902243).
- Install go using `apk add go`
- Clone this repository
- Run `make`
- Use opensend as normal, but skip device discovery
    - Device discovery does not work properly in iSH due to Alpine Linux
    - When running receiver, add `--skip-mdns`
    - When running sender, add `--send-to <IP>`
    - This applies bidirectionally
- Known issues
    - Opensend takes a while to become ready on iOS
 
### Ports to whitelist
- TCP 9797 for key exchange
- TCP 9898 for file transfer

### How does it work?
OpenSend uses a combination of 2048-bit RSA and AES GCM encryption. This is accomplished using golang's crypto/rsa and crypto/aes libraries. First, a shared AES key is generated. Then, an RSA keypair is generated. The RSA public key is then exchanged using TCP and golang's encoding/gob library.  The AES key is encrypted using the RSA public key
of the receiver. This key is then saved to a file. Next, the shared AES key is used to encrypt all the files in `~/.opensend`. To send the encrypted AES key, the sender first needs to discover the receiver. This is accomplished using zeroconf. After that, the sender starts an HTTP server with some custom functions to send the file index and key.
The receiver gets the index, files, and encrypted key from this server. Once it gets all the files, it sends a stop signal to the server and decrypts the shared key using its RSA private key. The resulting AES key is then used to decrypt all files in `~/.opensend`.