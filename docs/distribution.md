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
    └── sql/
        └── my-mod/
            ├── 20250129_120000_add_npc.sql
            ├── 20250129_120000_add_npc.rollback.sql
            ├── 20250129_130000_add_quest.sql
            └── 20250129_130000_add_quest.rollback.sql
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

## What's Included

| Content | Source | Install Location |
|---------|--------|------------------|
| DBC MPQ | Built from DBC database exports | `Data/patch-T.MPQ` |
| LuaXML MPQ | Built from mod LuaXML files | `Data/<locale>/patch-<locale>-T.MPQ` |
| World SQL | From `mods/<mod>/world_sql/` | Apply to world database |

**Note:** DBC SQL migrations are applied to the DBC database and exported to DBC files, which are then packaged into the client MPQ. The SQL files themselves are not distributed since the client needs the compiled DBC format.

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
