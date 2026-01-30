# Client Patcher

The client patcher applies binary patches to the WoW 3.3.5a (12340) executable. These are simple byte-level modifications that improve the client for private server development.

## Usage

```bash
# Apply all patches
thorium patch

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

## item-dbc-disabler

This patch makes the client request item data from the server instead of reading from its local `Item.dbc` file. This is useful for:

- **Dynamic item generation** - Create items on-the-fly from server scripts
- **Server-side item adjustments** - Modify item stats without redistributing client files
- **Custom items** - Add new items by inserting into `item_template` database table

Works with stock TrinityCore 3.3.5a (handles `CMSG_ITEM_QUERY_SINGLE` out of the box).

**Credits:**
- [TSWoW team](https://github.com/tswow/tswow) (`tswow-scripts/util/ClientPatches.ts`)
- [WoW 3.3.5 Patcher Custom Item Fix](https://www.wowmodding.net/files/file/283-wow-335-patcher-custom-item-fix/)

## Technical Details

Each patch modifies specific byte offsets in the WoW.exe binary. The patcher:

1. Creates a backup of the original executable (`.clean` suffix)
2. Verifies the source file MD5 hash matches the expected clean client
3. Applies all patches to the binary in memory
4. Writes the patched executable

The expected MD5 hash for a clean WoW 3.3.5a (12340) client is: `45892bdedd0ad70aed4ccd22d9fb5984`
