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
Alternatively, run the PowerShell server implementation:
```powershell
powershell.exe -ExecutionPolicy Bypass -File .\server\server.ps1
```
The PowerShell server supports the same `rdp2tcp` virtual channel handshake, yamux session handling, and SOCKS streams. It proxies SFTP streams to Windows OpenSSH `sftp-server.exe`, which the current Go client opens during startup; pass `-SftpServerPath C:\path\to\sftp-server.exe` if it is installed outside the default OpenSSH location. Reverse forwarding and Go-gob encoded local forwarding are not implemented in `server.ps1`.

3. Use the fwctrl from [forwardlib](https://github.com/Veids/forwardlib) repository to operate the client.
```bash
pipx install git+https://github.com/Veids/forwardlib
fwdctrl -h
```

## Usage examples

[forwardlib](https://github.com/Veids/forwardlib)
