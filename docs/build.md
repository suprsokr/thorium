# Build

The `thorium build` command is the core of the Thorium workflow. It performs a complete build pipeline that applies all mod changes, packages client files, and deploys server modifications.

## Requirements

Before building, [creating a workspace](init.md) and configuring <your-workspace>/config.json is required.

## Overview

`thorium build` executes an 8-step pipeline:

1. **Apply SQL Migrations** - Apply pending DBC and World database migrations
2. **Apply Binary Edits** - Patch the WoW client executable
3. **Apply Server Patches** - Apply Git patches to TrinityCore source
4. **Copy Mod Assets** - Copy asset files to the client directory
5. **Export Modified DBCs** - Generate DBC files from database tables
6. **Check LuaXML Modifications** - Detect modified interface files
7. **Deploy Scripts** - Copy C++ scripts to TrinityCore
8. **Package and Distribute** - Build MPQs and copy to client/server

## Basic Usage

```bash
# Build all mods (all 8 steps)
thorium build

# Build a specific mod only
thorium build --mod my-custom-items

# Build only specific components
thorium build dbc luaxml
thorium build binary scripts
thorium build world_sql
```

## Component Selection

You can build specific components by specifying them as positional arguments:

```bash
thorium build <component1> [component2] ...
```

**Available components:**
- `dbc` - DBC SQL migrations, export, and MPQ packaging
- `dbc_sql` - DBC SQL migrations only
- `world_sql` - World SQL migrations only
- `binary` - Binary edits to client executable
- `server-patches` - Server patches to TrinityCore source
- `assets` - Copy assets to client directory
- `luaxml` - LuaXML processing and MPQ packaging
- `scripts` - Deploy C++ scripts to TrinityCore

**Examples:**
```bash
# Build only DBC-related steps (migrations, export, packaging)
thorium build dbc

# Build only client-side changes (binary edits, assets, LuaXML)
thorium build binary assets luaxml
```

**Note:** When you specify components, only those components are built. If you don't specify any components, all steps are executed (unless skipped with flags).

## Flags

### Skip Options

Each step can be skipped individually:

```bash
--skip-dbc-sql         # Skip DBC SQL migrations (Step 1)
--skip-world-sql       # Skip World SQL migrations (Step 1)
--skip-binary-edits    # Skip binary edits (Step 2)
--skip-server-patches  # Skip server patches (Step 3)
--skip-assets          # Skip copying assets (Step 4)
--skip-export-dbc      # Skip DBC export (Step 5)
--skip-luaxml          # Skip LuaXML processing (Step 6)
--skip-scripts         # Skip script deployment (Step 7)
--skip-package         # Skip MPQ packaging (Step 8)
--skip-server-dbc      # Skip copying DBCs to server (part of Step 8)
```

**Example:**
```bash
# Only apply migrations and patches (no packaging)
thorium build --skip-export-dbc --skip-package

# Build everything except binary edits and assets
thorium build --skip-binary-edits --skip-assets
```

### Force Options

```bash
--force                    # Reapply everything (migrations, binary-edits, server-patches, assets, scripts)
--force-dbc-sql            # Reapply DBC SQL migrations only
--force-world-sql          # Reapply World SQL migrations only
--force-binary-edits       # Reapply binary edits only
--force-server-patches     # Reapply server patches only
--force-assets             # Recopy assets only
--force-scripts            # Redeploy scripts only
```

Force flags ignore tracking files and reapply items even if they've been applied before. Useful when:
- You manually reverted changes
- Tracking files are out of sync
- Testing changes to patches/edits
- Rebuilding after corruption

## Build Pipeline Details

### Step 1: Apply SQL Migrations

Applies pending migrations from `mods/<mod>/dbc_sql/` and `mods/<mod>/world_sql/`.

- **DBC migrations**: Requires DBC databases to be set up (see [dbc.md](dbc.md))
- **World migrations**: Applied to your TrinityCore world database
- **Tracking**: Migration history stored in marker files (`shared/applied/<mod>/<db_type>/*.applied`)
- **Skip**: Use `--skip-dbc-sql` to skip DBC migrations or `--skip-world-sql` to skip World migrations
- **Force**: Use `--force` to reapply all migrations, or `--force-dbc-sql` / `--force-world-sql` for specific types (rolls back and reapplies if rollback files exist)
- **Component**: Use `thorium build dbc_sql` or `thorium build world_sql` to build only specific migrations

If you haven't set up DBC databases yet and have DBC migrations, Thorium will prompt you to run `thorium init db`.

See [sql-migrations.md](sql-migrations.md) for details.

### Step 2: Apply Binary Edits

Discovers and applies binary patches from `mods/<mod>/binary-edits/*.json`.

- **Tracking**: `shared/binary_edits_applied.json`
- **Backup**: First run creates `Wow.exe.clean` backup
- **Verification**: Validates clean client MD5 hash
- **Requires**: `config.json` must have `wotlk.path` configured
- **Skip**: Use `--skip-binary-edits` to skip this step
- **Component**: Use `thorium build binary` to build only binary edits
- **Force**: Use `--force-binary-edits` to reapply

Example output:
```
┌──────────────────────────────────────────┐
│  Step 2: Applying Binary Edits           │
└──────────────────────────────────────────┘
  ✓ Clean WoW 3.3.5a client verified (MD5: 45892bde...)
[my-mod] Applying load-custom-dll.json...
  ✓ Applied load-custom-dll.json (2 patches)
Applied 1 new binary edit(s)
```

See [binary-edits.md](binary-edits.md) for details.

### Step 3: Apply Server Patches

Discovers and applies Git patches from `mods/<mod>/server-patches/*.patch`.

- **Tracking**: `shared/server_patches_applied.json`
- **Validation**: Runs `git apply --check` before applying
- **Requires**: `config.json` must have `trinitycore.source_path` configured
- **Skip**: Use `--skip-server-patches` to skip this step
- **Component**: Use `thorium build server-patches` to build only server patches
- **Force**: Use `--force-server-patches` to reapply
- **After**: You must rebuild TrinityCore for changes to take effect

Example output:
```
┌──────────────────────────────────────────┐
│  Step 3: Applying Server Patches         │
└──────────────────────────────────────────┘
[my-mod] Applying custom-opcodes.patch...
  ✓ Applied custom-opcodes.patch
Applied 1 new server patch(es)
  Note: Rebuild TrinityCore to apply changes
```

See [server-patches.md](server-patches.md) for details.

### Step 4: Copy Mod Assets

Copies files from `mods/<mod>/assets/` to your WoW client directory based on `assets/config.json`.

- **Tracking**: `shared/assets_applied.json` (with MD5 hashes)
- **Smart copying**: Only copies new or changed files
- **Requires**: `config.json` must have `wotlk.path` configured
- **Skip**: Use `--skip-assets` to skip this step
- **Component**: Use `thorium build assets` to build only assets
- **Force**: Use `--force-assets` to recopy everything

Example output:
```
┌──────────────────────────────────────────┐
│  Step 4: Copying Mod Assets              │
└──────────────────────────────────────────┘
[my-mod] Copied custom-texture.blp -> C:\WoW\Data\custom-texture.blp
Copied 1 asset(s) to client
```

See [assets.md](assets.md) for details.

### Step 5: Export Modified DBCs

Exports database tables that differ from the baseline to DBC files in `shared/dbc/dbc_out/`.

- **Comparison**: Uses checksums stored in the `dbc_checksum` table to detect modified tables
- **Checksum tracking**: Each DBC table has a checksum stored in the database; when migrations modify a table, its checksum changes
- **Efficiency**: Only exports tables whose checksums differ from the stored baseline (indicating modifications via SQL migrations)
- **Output**: `shared/dbc/dbc_out/*.dbc`
- **Skip**: Use `--skip-export-dbc` to skip this step
- **Component**: Use `thorium build dbc` to build DBCs (includes migrations, export, and packaging)
- **Requires**: DBC databases must be set up

Example output:
```
┌──────────────────────────────────────────┐
│  Step 5: Exporting Modified DBCs         │
└──────────────────────────────────────────┘
Exported 3 DBC table(s)
```

See [dbc.md](dbc.md#checksums) for more details.

### Step 6: Check LuaXML Modifications

Scans `mods/<mod>/luaxml/` for files that differ from the baseline in `shared/luaxml/luaxml_source/`.

- **Comparison**: Byte-by-byte comparison with source files
- **Includes**: UI files, addons, and any interface modifications
- **Output**: List of modified files for MPQ packaging
- **Skip**: Use `--skip-luaxml` to skip this step
- **Component**: Use `thorium build luaxml` to build only LuaXML (includes packaging)

Example output:
```
┌──────────────────────────────────────────┐
│  Step 6: Checking LuaXML Modifications   │
└──────────────────────────────────────────┘
[my-mod] Found 5 modified LuaXML file(s)
```

See [luaxml.md](luaxml.md) for details.

### Step 7: Deploy Scripts

Copies TrinityCore C++ scripts from `mods/<mod>/scripts/` to TrinityCore's scripts directory.

- **Tracking**: `shared/scripts_deployed.json` (with MD5 hashes)
- **Smart deployment**: Only deploys new or changed scripts
- **Requires**: `config.json` must have `trinitycore.scripts_path` configured
- **Skip**: Use `--skip-scripts` to skip this step
- **Component**: Use `thorium build scripts` to build only scripts
- **Force**: Use `--force-scripts` to redeploy everything
- **After**: You must rebuild TrinityCore for scripts to compile

Example output:
```
┌──────────────────────────────────────────┐
│  Step 7: Deploying Scripts               │
└──────────────────────────────────────────┘
[my-mod] Deployed custom_spell.cpp
Deployed 1 script(s)
  Note: Rebuild TrinityCore to compile scripts
```

See [scripts.md](scripts.md) for details.

### Step 8: Package and Distribute

Creates MPQ archives containing modified DBCs and LuaXML files, then copies them to the client and server.

- **DBC MPQ**: `patch-T.MPQ` in client `Data/` folder
- **LuaXML MPQ**: `patch-<locale>-T.MPQ` in client `Data/<locale>/` folder
- **Server DBCs**: Copies DBCs to server's `dbc/` directory (if configured)
- **Skip packaging**: Use `--skip-package`
- **Skip server copy**: Use `--skip-server-dbc`

Example output:
```
┌──────────────────────────────────────────┐
│  Step 8: Packaging and Distributing      │
└──────────────────────────────────────────┘
Copied 3 DBC file(s) to server
Created MPQ with DBC and LuaXML files and copied to client
```

See [distribution.md](distribution.md) for player distribution packages.

## Output Summary

After a successful build, you'll see a summary:

```
╔══════════════════════════════════════════╗
║           Build Complete!                ║
╚══════════════════════════════════════════╝

Output locations:
  Server DBCs: /home/peacebloom/server/bin/dbc
  Client DBC MPQ: C:\WoW\Data\patch-T.MPQ
  Client LuaXML MPQ: C:\WoW\Data\enUS\patch-enUS-T.MPQ
```

## What Gets Built

The build process only includes items that exist in your mods:

| Component | Source | Output | Requires |
|-----------|--------|--------|----------|
| DBC changes | `mods/<mod>/dbc_sql/*.sql` | `patch-T.MPQ`, server DBCs | DBC databases setup |
| World data | `mods/<mod>/world_sql/*.sql` | Applied to world database | World database |
| Binary edits | `mods/<mod>/binary-edits/*.json` | Modified `Wow.exe` | `wotlk.path` |
| Server patches | `mods/<mod>/server-patches/*.patch` | TrinityCore source changes | `trinitycore.source_path` |
| Assets | `mods/<mod>/assets/` | Files in client directory | `wotlk.path` |
| Interface mods | `mods/<mod>/luaxml/` | `patch-<locale>-T.MPQ` | `wotlk.path`  |
| Scripts | `mods/<mod>/scripts/*.cpp` | TrinityCore scripts | `trinitycore.scripts_path` |

## Tracking Files

Thorium tracks applied changes to avoid reapplying them unnecessarily:

| Tracker File | Tracks | Force Flag |
|-------------|--------|-----------|
| `shared/binary_edits_applied.json` | Binary edits | `--force-binary-edits` |
| `shared/server_patches_applied.json` | Server patches | `--force-server-patches` |
| `shared/assets_applied.json` | Asset files (with MD5) | `--force-assets` |
| `shared/scripts_deployed.json` | Deployed scripts (with MD5) | `--force-scripts` |
| SQL migrations table | Database migrations | N/A (use `rollback`) |

To reapply items, use the appropriate `--force-*` flag or edit the tracker file manually.

Environment variables like `${WOTLK_PATH}` are automatically expanded.

## After Building

### For Client Changes

If you modified DBCs, LuaXML, binary edits, or assets:

1. **Client files are ready** - The build automatically copied MPQs to your client. Client location is controlled in <your-workspace>/config.json
2. **Launch the game** - Just start WoW, changes will be loaded
3. **Test your changes** - Connect to your server and verify

### For Server Changes

If you applied server patches or deployed scripts:
1. Rebuild Trinity Core
2. Restart Server

## Comparison with `dist`

| Command | Purpose | Output | Use Case |
|---------|---------|--------|----------|
| `thorium build` | Development + deployment | Client files in WoW directory, server files in TC | Local testing and development |
| `thorium dist` | Player distribution | Zip file with MPQs and optional wow.exe | Sharing with players |

See [distribution.md](distribution.md) for more on creating player packages.
