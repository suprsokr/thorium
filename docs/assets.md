# Assets

Assets allow mods to copy files to the WoW client directory.

## Overview

- Assets are files in `mods/<mod>/assets/`
- Requires an `assets/config.json` to specify destinations
- Copied during `thorium build`
- Tracked by MD5 hash in `shared/assets_applied.json`
- Only copies if file has changed or doesn't exist at destination
- Destinations are relative to the WoW client directory

## Directory Structure

```
mods/my-mod/assets/
├── config.json           # Required: specifies where files go
├── ClientExtensions.dll  # Example: DLL to copy
└── custom-data/
    └── mydata.txt        # Example: nested file
```

## Configuration

Create `assets/config.json` to specify which files to copy and where:

```json
{
  "files": [
    {
      "source": "ClientExtensions.dll",
      "destination": "."
    },
    {
      "source": "custom-data/mydata.txt",
      "destination": "Data"
    }
  ]
}
```

### Fields

| Field | Description |
|-------|-------------|
| `files` | Array of file copy operations |
| `files[].source` | Path to source file, relative to `assets/` folder |
| `files[].destination` | Destination directory, relative to WoW client path |

### Destination Examples

| Destination | Result (assuming WoW at `/path/to/WoW/`) |
|-------------|------------------------------------------|
| `"."` | `/path/to/WoW/ClientExtensions.dll` |
| `"Data"` | `/path/to/WoW/Data/mydata.txt` |
| `"Interface/AddOns"` | `/path/to/WoW/Interface/AddOns/file.lua` |

## How It Works

1. During `thorium build`, each mod's `assets/config.json` is checked
2. For each file entry:
   - Source is read from `mods/<mod>/assets/<source>`
   - Destination directory is created if needed
   - File is copied to `<wotlk.path>/<destination>/<filename>`

## Configuration

Set the WoW client path in `config.json`:

```json
{
  "wotlk": {
    "path": "/path/to/WoW"
  }
}
```

Or use the environment variable:

```bash
export WOTLK_PATH=/path/to/WoW
```

## Example: DLL + Data Files

```
mods/custom-packets/assets/
├── config.json
├── ClientExtensions.dll
└── data/
    ├── opcodes.dat
    └── config.ini
```

**config.json:**

```json
{
  "files": [
    {
      "source": "ClientExtensions.dll",
      "destination": "."
    },
    {
      "source": "data/opcodes.dat",
      "destination": "Data"
    },
    {
      "source": "data/config.ini",
      "destination": "."
    }
  ]
}
```

**Result:**

```
/path/to/WoW/
├── ClientExtensions.dll
├── config.ini
└── Data/
    └── opcodes.dat
```

## Tracking

Assets are tracked by MD5 hash in `shared/assets_applied.json`:

```json
{
  "applied": [
    {
      "name": "custom-packets/ClientExtensions.dll",
      "md5": "a1b2c3d4e5f6...",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

If the source file's MD5 matches the tracked hash, the copy is skipped. Use `--force` to copy anyway:

```bash
thorium build --force
```

## Notes

- **Smart copying** - Only copies if file changed (based on MD5) or destination missing
- **Directories are created** - Destination directories are created automatically
- **No cleanup** - Thorium doesn't remove assets; delete them manually if needed

## Use Cases

| Use Case | Example |
|----------|---------|
| DLL injection | Copy `ClientExtensions.dll` for custom packets |
| Custom data files | Copy configuration or data files |
| Replacement resources | Copy modified textures/sounds (though MPQ is usually better) |

## Comparison with LuaXML

| Feature | Assets | LuaXML |
|---------|--------|--------|
| Goes into MPQ | No | Yes |
| Can be in subdirectories | Yes | Yes (must match Interface/ structure) |
| Tracking | By MD5 hash | Not tracked |
| Best for | DLLs, loose files | Addons, UI modifications |

For client-side Lua/XML modifications, use `luaxml/` instead - those files are packaged into an MPQ and loaded by the client's MPQ system.

## See Also

- [mods.md](mods.md) - Mod structure overview
- [binary-edits.md](binary-edits.md) - Patching Wow.exe
- [luaxml.md](luaxml.md) - Client-side Lua/XML in MPQs
