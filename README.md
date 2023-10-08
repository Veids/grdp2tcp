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
3. Use the fwctrl from [https://github.com/Veids/forwardlib](forwardlib) repository to operate the client.
```bash
pipx install git+https://github.com/Veids/forwardlib
fwdctrl -h
```

## Usage examples

[https://github.com/Veids/forwardlib](forwardlib)
