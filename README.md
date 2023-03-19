# grdp2tcp
Socks5 over RDP Virtual Channels

# Architecture
* server: Windows software that connects and opens a VC (virtual channel) via the WTS Windows API.
* client: Linux client that connects to the VC via XFreeRDP.

## Usage

1. Connect to the remote server via RDP using XFreeRDP or Remmina:
```bash
xfreerdp /v:{IP} /u:'{USER}' /p:'{PASS}' /rdp2tcp:$(pwd)/client
```
> For remmina, go to the advanced configuration tab and look for ***TCP redirection***.

2. Start server software on the Windows target.
```powershell
.\server.exe
```
3. Use the grdp2tcp.py tool to operate the client.

## Usage examples
### Start/stop socks server
```bash
#by default targets 127.0.0.1:1080
python grdp2tcp.py socks add
python grdp2tcp.py socks rm
#custom host:port
python grdp2tcp.py socks add -l 127.0.0.1:1090
python grdp2tcp.py socks rm -l 127.0.0.1:1090
```

### Start/stop reverse port forwarding
```bash
python grdp2tcp.py reverse add -l 127.0.0.1:8445 -r 127.0.0.1:8445
python grdp2tcp.py reverse rm -l 127.0.0.1:8445 -r 127.0.0.1:8445
```

### List configured endpoints
```bash
python grdp2tcp.py list
```
