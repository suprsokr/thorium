# Distribution

The `dist` command creates a player-ready distribution package containing client files (MPQs and optionally wow.exe). This is what you distribute to players who want to play on your modded server.

For mod source distribution (for other modders), host your mod on GitHub and use the `thorium get-mod <github-repo>` command.

## Usage

```bash
# Create a player distribution zip with MPQs
thorium dist

# Create a distribution for a specific mod
thorium dist --mod my-mod

# Skip including wow.exe (even if binary edits were applied)
thorium dist --no-exe

# Specify output path
thorium dist --output ./releases/v1.0.0.zip
```

**Note:** The command automatically checks `mods/shared/binary_edits_applied.json` and includes `wow.exe` from your WoTLK path if any binary edits have been applied (unless `--no-exe` is specified).

## Output Structure

The generated zip contains:

```
thorium_dist_20250201_143052.zip
├── README.txt              # Installation instructions for players
├── patch-T.MPQ             # DBC modifications (if built)
├── patch-enUS-T.MPQ        # LuaXML/interface modifications (if built)
└── wow.exe                 # Patched executable (if mod has binary edits)
```

**Note:** 
- MPQ files are only included if you've run `thorium build` first to generate them.
- `wow.exe` is automatically included if `mods/shared/binary_edits_applied.json` shows any applied binary edits.
- Use `--no-exe` flag to skip including wow.exe even if binary edits were applied.
- If no MPQ files exist, the command will inform you and not create a zip.

## Installation (for players)

Players should follow these steps to install your mod:

### 1. Backup

Backup the existing WoW 3.3.5a installation before proceeding.

### 2. Copy files

Extract the zip and copy files to the WoW folder:

```
WoW 3.3.5a/
├── wow.exe                 ← Copy if included in distribution
└── Data/
    ├── patch-T.MPQ         ← Copy to Data/
    └── enUS/
        └── patch-enUS-T.MPQ  ← Copy to Data/<locale>/
```

**Important:** Replace `enUS` with the client's locale (e.g., `enGB`, `deDE`, `frFR`).

### 3. Connect

Launch WoW and connect to your server. The client will load the custom MPQ files automatically.

## What's Included

| Content | Source | Install Location | Notes |
|---------|--------|------------------|-------|
| DBC MPQ | Built from DBC database exports | `Data/patch-T.MPQ` | Only if DBCs were modified |
| LuaXML MPQ | Built from mod LuaXML files + addons | `Data/<locale>/patch-<locale>-T.MPQ` | Only if addons/interface mods exist |
| wow.exe | From WoTLK path in config.json | Root folder | Automatically included if binary edits were applied |

**Notes:**
- DBC SQL migrations are temporarily applied to the `dbc_source` database, exported to DBC files, then rolled back to keep `dbc_source` pristine. This ensures only modified DBCs are included in the distribution. See [DBC Workflow](dbc.md#why-two-databases) for details.
- MPQ files are built by `thorium build` and stored at the paths specified in `config.json` (`output.dbc_mpq` and `output.luaxml_mpq`).
- `wow.exe` is automatically detected and included by checking `mods/shared/binary_edits_applied.json` for any applied binary edits. Use `--no-exe` to skip this.
- The SQL migration files themselves are not distributed since the client needs the compiled DBC format.

## Typical Workflow

A typical release workflow for sharing your mod with players:

```bash
# 1. Build everything (applies migrations, exports DBCs, creates MPQs, patches binaries)
thorium build

# 2. Test locally to ensure everything works

# 3. Create player distribution package
thorium dist --output ./releases/v1.0.0.zip

# 4. Share the zip with players (upload to website, Discord, etc.)
```

**How DBCs are packaged:** The `dist` command temporarily applies your mod's DBC migrations to the `dbc_source` database, exports only the modified DBCs, then rolls back the migrations to restore `dbc_source` to its pristine state. This ensures the distribution package only contains DBCs that were actually modified by your mod. See [DBC Workflow](dbc.md#why-two-databases) for more details.

## Mod Source Distribution (for modders)

The `dist` command is for **player distributions only**. If you want to share your mod source code with other modders, you should:

1. Host your mod on GitHub (the mod folder structure with scripts/, assets/, luaxml/, etc.)
2. Other modders can clone/download your mod
3. They run `thorium build` in their workspace to compile and apply your mod

Future versions will include `thorium get-mod <github-repo>` to automate fetching and installing mods from GitHub.
