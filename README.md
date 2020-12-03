# OpenSend
## Send files between systems quickly and securely

### How does it work?
OpenSend uses a combination of 2048-bit RSA and AES GCM encryption. This is accomplished using golang's crypto/rsa and crypto/aes libraries. First, a shared AES key is generated. Then, an RSA keypair is generated. The AES key is encrypted using the RSA public key
of the receiver. This key is then saved to a file. Next, the shared AES key is used to encrypt all the files in `~/.opensend`. To send the key, the sender first needs to discover the receiver. This is accomplished using mDNS. The key is then exchanged using a TCP socket and golang's encoding/gob library. After that, the sender starts an HTTP server with some custom functions to send the file index and key. The receiver gets the index, files, and encrypted key from this server. Once it gets all the files, it sends a stop signal to the server and decrypts the shared key using its RSA private key. The resulting key is then used to decrypt all files in `~/.opensend`.

### Usage

#### Receiver
- Use `opensend -r` to start the receiver

### Sender
- Use `opensend -s -t <type> -d <data>`
- Type can either be `url` or `file`
- If the type is `url`, the data should be a URL
- If the type is `file`, the data should be a file path

### Building
- This project uses go modules, so building is easy
- First, go 1.15+ must be installed
- Then, run `go build` inside the project's directory.
- This will get the dependencies and compile all the files.

### Installing
To install, simply follow the building instructions and then run
- Linux: `sudo install -Dm755 opensend /usr/local/bin`
- macOS: `sudo install -m755 opensend /usr/local/bin`

