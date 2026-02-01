# Client Patcher

The client patcher applies binary patches to the WoW 3.3.5a (12340) executable. These are simple byte-level modifications that improve the client for private server development.

## Usage

```bash
# Apply all patches (uses wotlk.path from config.json)
thorium patch

# Apply patches to a specific exe (no config.json needed)
thorium patch /path/to/WoW.exe

# List available patches
thorium patch --list

# Preview what would be applied
thorium patch --dry-run

# Restore original client from backup
thorium patch --restore
```

A backup (`WoW.exe.clean`) is automatically created before the first patch is applied.

## Available Patches

| Patch | Description | Why It Helps |
|-------|-------------|--------------|
| `large-address-aware` | Sets the LAA flag in the PE header, allowing the 32-bit client to use up to 4GB of RAM instead of 2GB. | Prevents out-of-memory crashes when using heavy addons, custom content, or high-resolution textures. Essential for modded clients. |
| `allow-custom-gluexml` | Bypasses signature checks on GlueXML files (login screen UI). | Allows custom login screens, server selection UI, and other glue screen modifications without triggering MPQ signature errors. |
| `wowtime-fix` | Fixes calendar and date calculation bugs that cause crashes or incorrect dates after January 1, 2031. | Future-proofs your server. Without this patch, clients will crash or show wrong dates starting in 2031. |
| `item-dbc-disabler` | Disables client-side Item.dbc lookups, forcing the client to request item data from the server. | Enables dynamic item generation and server-side item adjustments without client file changes. |
| `custom-packets` | Injects ClientExtensions.dll to enable custom packet protocol. | Enables bidirectional communication between client addons and server for custom features. See details below. |

## item-dbc-disabler

This patch makes the client request item data from the server instead of reading from its local `Item.dbc` file. This is useful for:

- **Dynamic item generation** - Create items on-the-fly from server scripts
- **Server-side item adjustments** - Modify item stats without redistributing client files
- **Custom items** - Add new items by inserting into `item_template` database table

Works with stock TrinityCore 3.3.5a (handles `CMSG_ITEM_QUERY_SINGLE` out of the box).

**Credits:**
- [TSWoW team](https://github.com/tswow/tswow) (`tswow-scripts/util/ClientPatches.ts`)
- [WoW 3.3.5 Patcher Custom Item Fix](https://www.wowmodding.net/files/file/283-wow-335-patcher-custom-item-fix/)

## Custom Packets

The `custom-packets` patch enables bidirectional communication between client addons and server scripts using custom opcodes. This allows you to build features like custom UI, real-time data sync, and server-driven client behavior.

**Important:** This feature requires BOTH client patches AND server patches:
- **Client:** `thorium patch` (applies this patch + installs ClientExtensions.dll)
- **Server:** `thorium patch-server apply custom-packets` (patches TrinityCore source)

### How It Works

This patch injects `ClientExtensions.dll` into the WoW client at startup. The DLL:

1. **Hooks client networking functions** using Microsoft Detours
2. **Intercepts custom opcodes** (0x102 for server→client, 0x51F for client→server)
3. **Provides Lua API** for addons to send/receive custom packets
4. **Handles packet fragmentation** for large messages (up to 8MB)

When you run `thorium patch`, it:
- Patches `WoW.exe` to load `ClientExtensions.dll` alongside `d3d9.dll`
- Copies `ClientExtensions.dll` to your WoW directory

When you run `thorium patch-server apply custom-packets`, it:
- Patches TrinityCore source to register opcode 0x51F
- Adds `OnCustomPacketReceive` hook to `ServerScript` for your handlers
- Requires rebuilding TrinityCore after applying

When you run `thorium init`, a `CustomPackets` addon is automatically created that provides the client-side Lua API.

### Credits

- **ClientExtensions.dll**: Based on [TSWoW](https://github.com/tswow/tswow) client-extensions (MIT License)
- Source and releases: https://github.com/suprsokr/wotlk-custom-packets

See [custom-packets.md](custom-packets.md) for complete documentation including:

- Lua API for sending/receiving packets
- Server-side C++ handlers  
- Data types and packet structure
- Setup guide and best practices

## Technical Details

Most patches modify specific byte offsets in the WoW.exe binary. The `custom-packets` patch injects a DLL. The patcher:

1. Creates a backup of the original executable (`.clean` suffix)
2. Verifies the source file MD5 hash matches the expected clean client
3. Applies all patches to the binary in memory
4. Writes the patched executable
5. For `custom-packets`: Copies `ClientExtensions.dll` to the WoW directory

The expected MD5 hash for a clean WoW 3.3.5a (12340) client is: `45892bdedd0ad70aed4ccd22d9fb5984`

### DLL Injection Details

The `custom-packets` patch modifies three memory locations in `WoW.exe`:

1. **0x56DE90**: Redirects the `LoadLibraryA` call to a code cave
2. **0x3738b8**: Code cave that loads both `d3d9.dll` and `ClientExtensions.dll`
3. **0x5e2a71**: String "ClientExtensions.dll\0" in unused memory

This is a safe, non-intrusive approach that loads the DLL during the game's normal startup sequence.
