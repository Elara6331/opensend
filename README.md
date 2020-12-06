# OpenSend
## Send files between systems quickly and securely

### Usage

#### Receiver
- Use `opensend -r` to start the receiver

#### Sender
- Use `opensend -s -t <type> -d <data>`
- `type` can be
    - `url`
    - `file`
    - `dir`
- `data` can be
    - A website URL
    - A file path
    - A directory path
- Example: `opensend -s -t url -d "https://google.com"`
- Example: `opensend -s -t file -d ~/file.txt`
- Example: `opensend -s -t dir -d /home/user`

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