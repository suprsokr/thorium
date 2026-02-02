# Status

The `thorium status` command shows the current state of your workspace, including all mods, their migrations, and LuaXML files.

## Overview

`thorium status` provides a quick overview of:
- **Workspace location** - Where your Thorium workspace is located
- **Mods** - List of all mods in your workspace
- **Migration status** - Which migrations are applied vs pending for each mod
- **LuaXML files** - Count of LuaXML files in each mod

## Usage

```bash
# Show status for all mods
thorium status

# Show status for a specific mod
thorium status --mod my-mod
```

### Flags

- `--mod <name>` - Show status for a specific mod only

## Output Format

The status command displays information in a structured format:

**All mods:**
```
=== Thorium Status ===

Workspace: /path/to/workspace

Found 2 mod(s):
  - my-custom-items
  - my-custom-spells

=== my-custom-items ===
  DBC Migrations:
    [applied] 20250129_120000_add_custom_spell.sql
    [pending] 20250130_140000_update_spell.sql
  World Migrations:
    [applied] 20250129_130000_add_custom_npc.sql
  LuaXML Files: 5

=== my-custom-spells ===
  DBC Migrations:
    [pending] 20250131_100000_add_fireball.sql
  LuaXML Files: 0
```

**Single mod (with `--mod` flag):**
```
=== Thorium Status ===

Workspace: /path/to/workspace

=== my-custom-items ===
  DBC Migrations:
    [applied] 20250129_120000_add_custom_spell.sql
    [pending] 20250130_140000_update_spell.sql
  World Migrations:
    [applied] 20250129_130000_add_custom_npc.sql
  LuaXML Files: 5
```

## What Gets Shown

### Workspace Information

The command first displays your workspace root path, which helps verify you're in the correct directory.

### Mods List

All mods found in `mods/` are listed. Mods are discovered by scanning the `mods/` directory for subdirectories (excluding hidden directories starting with `.`).

### Migration Status

For each mod, the command shows:

- **DBC Migrations** - SQL files in `mods/<mod>/dbc_sql/`
  - `[applied]` - Migration has been applied (`.applied` marker file exists)
  - `[pending]` - Migration exists but hasn't been applied yet

- **World Migrations** - SQL files in `mods/<mod>/world_sql/`
  - `[applied]` - Migration has been applied (`.applied` marker file exists)
  - `[pending]` - Migration exists but hasn't been applied yet

Migrations are detected by looking for `.sql` files (excluding `.rollback.sql` files) and checking for corresponding `.applied` marker files in `shared/migrations_applied/<mod>/<db_type>/`.

### LuaXML Files

The count of LuaXML files in each mod's `luaxml/` directory. This includes all files recursively, excluding hidden files (those starting with `.`).

## Use Cases

### Check Migration Status

Quickly see which migrations need to be applied:

```bash
thorium status
# Look for [pending] migrations
thorium build dbc_sql world_sql  # Apply pending migrations
```

### Verify Mod Structure

Ensure your mods are properly structured and discoverable:

```bash
thorium status
# Verify all expected mods are listed
```

### Before Building

Check what will be included in your next build:

```bash
thorium status
# Review migrations, LuaXML files
thorium build
```

## Empty Workspace

If no mods are found, the command will display:

```
=== Thorium Status ===

Workspace: /path/to/workspace

No mods found in /path/to/workspace/mods
Run 'thorium create-mod <name>' to create one.
```

## Notes

- **Migration tracking**: Applied migrations are tracked via marker files in `shared/migrations_applied/`. If you manually delete these files, migrations will show as `[pending]` even if they've been applied to the database.

- **Rollback files**: Rollback files (`.rollback.sql`) are not shown in the status output, only the main migration files.

- **Hidden files**: Hidden files and directories (starting with `.`) are excluded from counts and listings.

- **Mod filtering**: Use `--mod` to show status for a single mod. When filtering, the mods list is not displayed.

## Related Commands

- `thorium build` - Apply pending migrations and build all components
- `thorium build dbc_sql world_sql` - Apply only migrations
- `thorium rollback` - Undo applied migrations
