# Installation

## From Source

```bash
git clone --recursive https://github.com/suprsokr/thorium
cd thorium
make # `make install` for system-wide installation
```

This builds a single `thorium` binary (~10MB) with everything included.

## Requirements

**Note:** Ubuntu 24.04 is officially supported and tested. Other distributions and versions are untested and may require adjustments.

### Server Prerequisites

- A Trinity Core 3.3.5 server installed and functioning
- MySQL 8.0 (see [Trinity Core requirements](https://trinitycore.info/en/install/requirements) for full details)

### Build Requirements

**Ubuntu 24.04 (Officially Supported):**
```bash
sudo apt update
sudo apt install golang-go cmake g++ zlib1g-dev libbz2-dev
```

**macOS:**
```bash
brew install go@1.23 cmake
# Xcode command line tools provide clang, zlib, bzip2
xcode-select --install
```

**Other Ubuntu/Debian versions (Untested):**
```bash
sudo apt update
sudo apt install golang-go cmake g++ zlib1g-dev libbz2-dev
# Requires Go 1.21 or later
```

**Fedora/RHEL (Untested):**
```bash
sudo dnf install golang cmake gcc-c++ zlib-devel bzip2-devel
# Requires Go 1.21 or later. Verify with: go version
```

**Arch Linux (Untested):**
```bash
sudo pacman -S go cmake gcc zlib bzip2
# Requires Go 1.21 or later. Verify with: go version
```

**Windows (with MSYS2) (Untested):**
```bash
pacman -S mingw-w64-x86_64-go mingw-w64-x86_64-cmake mingw-w64-x86_64-gcc mingw-w64-x86_64-zlib mingw-w64-x86_64-bzip2
# Requires Go 1.21 or later. Verify with: go version
```

## Configuration

After installation, initialize a workspace:

```bash
thorium init
```

This creates a `config.json` file. Edit it to configure your paths and database connections:

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
    "dbc_path": "${TC_SERVER_PATH}/data/dbc"
  }
}
```

### Configuration Options

- **wotlk.path**: Path to your WoW 3.3.5 client directory. Can use `${WOTLK_PATH}` environment variable.
- **wotlk.locale**: Client locale (e.g., `enUS`, `enGB`, `deDE`).
- **databases.dbc**: Connection settings for the DBC database (where DBC data is stored as SQL tables).
- **databases.world**: Connection settings for your TrinityCore world database.
- **server.dbc_path**: Path where TrinityCore expects DBC files. Can use `${TC_SERVER_PATH}` environment variable.

Environment variables like `${WOTLK_PATH}` and `${TC_SERVER_PATH}` are automatically expanded.

### Setting Up the DBC Database

Before using Thorium, you need to create and populate the DBC database:

```bash
# Extract DBCs from your WoW client
thorium extract --dbc

# Import DBCs into the database
# (This step may vary depending on your setup)
```

The DBC database stores all client data files (items, spells, NPCs, etc.) as SQL tables, allowing you to edit them with SQL migrations.
