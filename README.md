# gomix

A simple MIDI to DBus mixer built Golang. It's hardcoded for my pipewire setup.

## Installation

Clone the repository and copy the binary to /usr/local/bin. Then, copy the systemd service file to $HOME/.local/share/systemd/user and enable it.

```bash
git clone https://github.com/pernydev/gomix.git
sudo cp gomix/gomix /usr/local/bin
sudo chmod +x /usr/local/bin/gomix
cp gomix/gomix.service $HOME/.local/share/systemd/user
systemctl --user enable gomix.service
systemctl --user start gomix.service
```
