# gomix

A simple MIDI to Pipewire mixer built Golang. It's hardcoded for my pipewire setup.

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

![IMG_20240107_142516471](https://github.com/pernydev/gomix/assets/83672513/86448be0-9aea-4fd4-bcdc-d00ea6de2660)
![image](https://github.com/pernydev/gomix/assets/83672513/f174fccd-27f8-4442-b941-a82ef69006ca)
![image](https://github.com/pernydev/gomix/assets/83672513/f89907bc-2814-4209-9c56-f4a46d72f8eb)
