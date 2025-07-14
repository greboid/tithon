# <img alt="Tithon" src="/frontend/icon.png" height="50px"> Tithon - Modern Desktop IRC Client

Tithon is a modern, cross-platform IRC client that combines the power of IRC with a clean desktop interface. Built with Go and Electron, it leverages IRCv3 features to provide a smooth messaging experience while staying true to the IRC protocol.

## Features

- **IRCv3 Support**: Lots of support for modern IRC features and capabilities
- **Cross-Platform**: Available for Linux, Windows, and macOS
- **Theme Support**: Light, dark, and auto themes, but also a user.css file that can be used for extensive customization
- **File Upload**: Built-in support for file sharing via upload URLs

## Installation

### Arch Linux
Tithon is available in the AUR:
```bash
paru -S tithon
```

### Debian/Ubuntu
Download the latest `.deb` package from the [Releases page](https://github.com/greboid/tithon/releases/latest) and install:
```bash
sudo dpkg -i tithon_*.deb
```

### Other Platforms
Download the appropriate package for your platform from the [Releases page](https://github.com/greboid/tithon/releases/latest).

## Quick Start

1. **Launch Tithon**
   - Click the settings icon in the bottom left
   - Click select Servers
2. **Configure your server**:
   - Enter the server hostname (e.g., `irc.libera.chat`)
   - Set your nickname
   - Enable TLS for secure connections (recommended)
   - Configure SASL authentication if required
3. **Join channels** using `/join #channelname`
4. **Start chatting!**

## Configuration

Tithon stores its configuration in YAML format at:
- **Linux**: `$XDG_CONFIG_HOME/tithon/config.yaml` (usually `~/.config/tithon/config.yaml`)
- **Windows**: `%APPDATA%\tithon\config.yaml`
- **macOS**: `~/Library/Application Support/tithon/config.yaml`

### Notification Configuration

These are only available by editing the config file directly, the client will need to be shut to do this

```yaml
notifications:
  triggers:
    - network: "libera"            # Network name (optional)
      source: "#channel"           # Channel or user (optional)
      nick: "YourNick"            # Trigger on mentions (optional)
      message: "keyword"          # Trigger on keywords (optional)
      sound: true                 # Play notification sound
      popup: true                 # Show desktop notification
```

## File Uploads

If you're using Soju, you can enable filehost support and this will automatically be picked up, otherwise (or instead of) you can configure file uploads by setting the upload URL in your configuration:

1. Set `upload_url` to your upload service endpoint
2. Optionally configure `upload_api_key` for authentication
3. Optionally set `upload_method` this defaults to `POST`
4. Click the upload button or paste files

## License

Tithon is released under the MIT License. See the LICENSE file for details.
