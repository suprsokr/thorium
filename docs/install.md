# Installation

## Pre-built Binaries (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/suprsokr/thorium/releases):

```bash
# Linux x64
curl -L https://github.com/suprsokr/thorium/releases/latest/download/thorium-linux-amd64 -o thorium
chmod +x thorium
sudo mv thorium /usr/local/bin/

# Linux ARM64 (e.g., Raspberry Pi, AWS Graviton)
curl -L https://github.com/suprsokr/thorium/releases/latest/download/thorium-linux-arm64 -o thorium

# macOS Intel
curl -L https://github.com/suprsokr/thorium/releases/latest/download/thorium-darwin-amd64 -o thorium

# macOS Apple Silicon
curl -L https://github.com/suprsokr/thorium/releases/latest/download/thorium-darwin-arm64 -o thorium

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/suprsokr/thorium/releases/latest/download/thorium-windows-amd64.exe -OutFile thorium.exe
```

## From Source

Requires Go 1.22 or later.

```bash
# Install directly
go install github.com/suprsokr/thorium/cmd/thorium@latest

# Or clone and build
git clone https://github.com/suprsokr/thorium.git
cd thorium
go build -o thorium ./cmd/thorium
sudo mv thorium /usr/local/bin/
```

### Cross-Compile

Build for other platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o thorium-linux-amd64 ./cmd/thorium

# macOS
GOOS=darwin GOARCH=arm64 go build -o thorium-darwin-arm64 ./cmd/thorium

# Windows
GOOS=windows GOARCH=amd64 go build -o thorium-windows-amd64.exe ./cmd/thorium
```

## Requirements

### Server Prerequisites

- A TrinityCore 3.3.5 server installed and functioning
- MySQL 8.0 (see [TrinityCore requirements](https://trinitycore.info/en/install/requirements))

### Build Requirements

Just Go 1.22 or later. Thorium is pure Go with no C dependencies.

```bash
# Ubuntu/Debian
sudo apt install golang-go

# macOS
brew install go

# Fedora/RHEL
sudo dnf install golang

# Arch Linux
sudo pacman -S go

# Windows
# Download from https://go.dev/dl/
```

Verify your Go version: `go version`

## Configuration

After installation, initialize a workspace:

```bash
thorium init
```

This creates the workspace directory structure and a `config.json` file. See [init.md](init.md) for detailed information on initialization.

Edit `config.json` to configure your paths and database connections, or use [Peacebloom](https://github.com/suprsokr/peacebloom) to manage your environment with Docker:

```json
{
  "wotlk": {
    "path": "${WOTLK_PATH}",
    "locale": "enUS"
  },
  "databases": {
    "dbc": {
      "user": "root",
      "host": "127.0.0.1",
      "port": "3306",
      "name": "dbc"
    },
    "world": {
      "user": "trinity",
      "password": "trinity",
      "host": "127.0.0.1",
      "port": "3306",
      "name": "world"
    }
  },
  "server": {
    "dbc_path": "${TC_SERVER_PATH:-/home/peacebloom/server}/bin/dbc"
  }
}
```

### Configuration Options

See [Configuration](config.md) for complete documentation of all `config.json` parameters.

Environment variables like `${WOTLK_PATH}` and `${TC_SERVER_PATH}` are automatically expanded. See [DBC Workflow](dbc.md#just-in-time-setup) for when and how to set up DBC databases.
