# Distribution

The `dist` command creates a distributable package containing everything needed to deploy your mods to a client and server.

## Usage

```bash
# Create a zip of all mods
thorium dist

# Create a zip for a specific mod
thorium dist --mod my-mod

# Specify output path
thorium dist --output ./releases/my-release.zip
```

## Output Structure

The generated zip contains:

```
thorium_dist_20250129_143052.zip
├── README.txt              # Installation instructions
├── client/
│   ├── patch-T.MPQ             # DBC modifications (goes in Data/)
│   └── patch-enUS-T.MPQ        # LuaXML/interface modifications (goes in Data/enUS/)
└── server/
    ├── sql/
    │   └── my-mod/
    │       ├── 20250129_120000_add_npc.sql
    │       ├── 20250129_120000_add_npc.rollback.sql
    │       ├── 20250129_130000_add_quest.sql
    │       └── 20250129_130000_add_quest.rollback.sql
    └── scripts/
        └── my-mod/
            ├── spell_fire_blast.cpp
            └── npc_custom_vendor.cpp
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

## What's Included

| Content | Source | Install Location |
|---------|--------|------------------|
| DBC MPQ | Built from DBC database exports | `Data/patch-T.MPQ` |
| LuaXML MPQ | Built from mod LuaXML files + addons | `Data/<locale>/patch-<locale>-T.MPQ` |
| World SQL | From `mods/<mod>/world_sql/` | Apply to world database |
| Scripts | From `mods/<mod>/scripts/` | Copy to TrinityCore Custom/, rebuild |

**Notes:**
- DBC SQL migrations are applied to the DBC database and exported to DBC files, which are then packaged into the client MPQ. The SQL files themselves are not distributed since the client needs the compiled DBC format.
- Scripts are distributed as C++ source files. Recipients must copy them to their TrinityCore source and rebuild the server.
- The LuaXML MPQ includes the `CustomPackets` addon (created during `thorium init`) and any addons created with `thorium create-addon`.

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
