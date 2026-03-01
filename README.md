# Atlas

A terminal email client built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- IMAP/SMTP with native TLS
- Folder browsing, message reading, and composing
- Reply, reply-all, and forward
- Interactive setup wizard
- TOML config with environment variable overrides

## Install

### Quick install (Linux / macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/chazzychouse/atlas/main/install.sh | sh
```

### Quick install (Windows PowerShell)

```powershell
irm https://raw.githubusercontent.com/chazzychouse/atlas/main/install.ps1 | iex
```

### From source

```sh
go install github.com/chazzychouse/atlas@latest
```

### Manual download

Grab the latest archive from [Releases](https://github.com/chazzychouse/atlas/releases), then:

```sh
tar -xzf atlas_linux_amd64.tar.gz
sudo mv atlas /usr/local/bin/
```

## Setup

Run the interactive setup wizard to create your config:

```sh
atlas setup
```

This writes a TOML config to `~/.config/atlas/config.toml`.

### Environment variables

Any setting can be overridden with an environment variable:

| Variable             | Description       |
| -------------------- | ----------------- |
| `ATLAS_IMAP_HOST`    | IMAP server host  |
| `ATLAS_IMAP_PORT`    | IMAP server port  |
| `ATLAS_IMAP_USER`    | IMAP username     |
| `ATLAS_IMAP_PASS`    | IMAP password     |
| `ATLAS_SMTP_HOST`    | SMTP server host  |
| `ATLAS_SMTP_PORT`    | SMTP server port  |
| `ATLAS_SMTP_USER`    | SMTP username     |
| `ATLAS_SMTP_PASS`    | SMTP password     |
| `ATLAS_FROM_NAME`    | Sender name       |
| `ATLAS_FROM_EMAIL`   | Sender email      |

## Usage

```sh
atlas
```

## License

MIT
