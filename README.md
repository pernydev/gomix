# gomix
A simple MIDI to DBus mixer built Golang. It's hardcoded for my pipewire setup.

## Installation
Clone the repository and copy the binary to /usr/local/bin. Then, copy the systemd service file to /etc/systemd/system and enable it.

```bash
git clone https://github.com/pernydev/gomix.git
sudo cp gomix/gomix /usr/local/bin
sudo cp gomix/gomix.service /etc/systemd/system
sudo systemctl enable gomix.service
sudo systemctl start gomix.service
```