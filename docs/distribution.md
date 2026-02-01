# Distribution

The `dist` command creates a distributable package containing everything needed to deploy your mods to a client and server.

## Usage

```bash
# Create a zip of all mods (client MPQs + server SQL)
thorium dist

# Create a zip for a specific mod
thorium dist --mod my-mod

# Client-only distribution (just MPQs, no server SQL)
thorium dist --client-only

# Include patched wow.exe in distribution
thorium dist --client-only --include-exe

# Specify output path
thorium dist --output ./releases/my-release.zip
```

### Distribution Types

| Flags | Contents | Use Case |
|-------|----------|----------|
| (default) | Client MPQs + Server SQL | Server admins deploying full mod |
| `--client-only` | Client MPQs only | Players connecting to modded server |
| `--include-exe` | Also include patched wow.exe | Players needing client patches |

## Output Structure

The generated zip contains:

```
thorium_dist_20250129_143052.zip
├── README.txt              # Installation instructions
├── client/
│   ├── patch-T.MPQ             # DBC modifications (goes in Data/)
│   ├── patch-enUS-T.MPQ        # LuaXML/interface modifications (goes in Data/enUS/)
│   └── ClientExtensions.dll    # Custom packets DLL (if custom-packets enabled)
└── server/
    ├── sql/
    │   └── my-mod/
    │       ├── 20250129_120000_add_npc.sql
    │       ├── 20250129_120000_add_npc.rollback.sql
    ├── scripts/
    │   └── my-mod/
    │       ├── spell_fire_blast.cpp
    └── patches/
        └── custom-packets.patch  # TrinityCore source patch (if custom-packets enabled)
```

## Installation (for recipients)

### Client

Copy MPQ files to the appropriate WoW `Data/` directories:

```
WoW 3.3.5a/
└── Data/
    ├── patch-T.MPQ         ← DBC patch (Data/)
    └── enUS/
        └── patch-enUS-T.MPQ  ← LuaXML patch (Data/<locale>/)
```

**Note:** Replace `enUS` with your client's locale (e.g., `enGB`, `deDE`, `frFR`).

### Server

#### 1. Apply SQL migrations

Run the SQL files against the world database in order:

```bash
# Apply migrations (skip .rollback.sql files)
mysql -u root -p world < server/sql/my-mod/20250129_120000_add_npc.sql
mysql -u root -p world < server/sql/my-mod/20250129_130000_add_quest.sql
```

To undo changes, run the rollback files in reverse order:

```bash
mysql -u root -p world < server/sql/my-mod/20250129_130000_add_quest.rollback.sql
mysql -u root -p world < server/sql/my-mod/20250129_120000_add_npc.rollback.sql
```

#### 2. Install scripts (if included)

If the mod includes C++ scripts, copy them to your TrinityCore source:

```bash
# Copy scripts to TrinityCore Custom directory
cp -r server/scripts/my-mod/* /path/to/TrinityCore/src/server/scripts/Custom/

# Rebuild TrinityCore
cd /path/to/TrinityCore/build
make -j$(nproc)

# Restart server
systemctl restart worldserver  # or your restart method
```

Scripts are compiled into the server binary, so a rebuild is required.

#### 3. Apply server patches (if included)

If the mod includes server patches (e.g., for custom packets support):

```bash
# Navigate to TrinityCore source
cd /path/to/TrinityCore

# Apply the patch
git apply /path/to/server/patches/custom-packets.patch

# Rebuild TrinityCore
cd build
make -j$(nproc)
```

To revert a patch:
```bash
git apply -R /path/to/server/patches/custom-packets.patch
```

#### 4. Install ClientExtensions.dll (if included)

If the mod uses custom packets, copy `ClientExtensions.dll` to the WoW directory:

```bash
cp client/ClientExtensions.dll /path/to/WoW/
```

**Note:** Players also need a patched `WoW.exe` that loads this DLL. Either include the patched exe with `--include-exe`, or have players run `thorium patch` themselves.

## What's Included

| Content | Source | Install Location |
|---------|--------|------------------|
| DBC MPQ | Built from DBC database exports | `Data/patch-T.MPQ` |
| LuaXML MPQ | Built from mod LuaXML files + addons | `Data/<locale>/patch-<locale>-T.MPQ` |
| World SQL | From `mods/<mod>/world_sql/` | Apply to world database |
| Scripts | From `mods/<mod>/scripts/` | Copy to TrinityCore Custom/, rebuild |
| ClientExtensions.dll | Embedded in thorium | `WoW.exe` directory (client) |
| Server patches | From thorium assets | Apply to TrinityCore source, rebuild |

**Notes:**
- DBC SQL migrations are applied to the DBC database and exported to DBC files, which are then packaged into the client MPQ. The SQL files themselves are not distributed since the client needs the compiled DBC format.
- Scripts are distributed as C++ source files. Recipients must copy them to their TrinityCore source and rebuild the server.
- The LuaXML MPQ includes the `CustomPackets` addon (created during `thorium init`) and any addons created with `thorium create-addon`.

### Custom Packets Files

If your mod uses custom packets (`extensions.custom_packets.enabled: true` in config.json), the distribution includes:

**Client-side:**
- `ClientExtensions.dll` - Copy to WoW directory (next to `WoW.exe`)
- The patched `WoW.exe` (if `--include-exe` flag used)

**Server-side:**
- `patches/custom-packets.patch` - Apply to TrinityCore source with `git apply`

See [custom-packets.md](custom-packets.md) for full setup instructions.

## Workflow

A typical release workflow:

```bash
# 1. Build everything (applies migrations, exports DBCs, creates MPQs)
thorium build

# 2. Test locally

# 3. Create distribution package
thorium dist --output ./releases/v1.0.0.zip

# 4. Share the zip with players/server admins
```
