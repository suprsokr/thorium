# Thorium

A unified workflow for creating WoW 3.3.5 server and client-side mods.

## Why Thorium?

**Works with stock TrinityCore.** No forked core required. Your mods are portable and easy to share.

**Organized changes into mods.** Each mod is a self-contained collection of dbc edits, world database edits, interface edits and c++ scripts.

**Handles the boring stuff.** Packaging and distribution. A single command to see your mod in action locally or to get it bundled for your end users. Focus on your mod, not the pipeline.

**LLM-friendly.** Thorium's SQL, C++ scripts and Lua-based files lets AI assistants help you build mods without getting lost in binary formats or complex toolchains.

## Features

### DBC Editing via SQL
Modify client data files—items, spells, creatures, maps—using familiar SQL syntax. Thorium imports DBCs into MySQL tables, lets you edit with migrations, then exports and distributes automatically to both client and server.

### World Database Migrations
Same migration workflow for TrinityCore's world database. Version-controlled, with apply and rollback for every change.

### LuaXML & Custom Addons
Extract and modify client interface files. Create custom addons that get packaged into MPQs automatically. Your UI changes distribute with a single command.

### Custom Packets
Build features that need real-time client-server communication. Send custom data between your addons and server scripts with a simple Lua API. See [docs/custom-packets.md](docs/custom-packets.md).

### C++ Script Scaffolding
Generate TrinityCore script templates for spells, creatures, and packet handlers. Thorium sets up the boilerplate scripts and distributes into your Trinity Core server so you can focus on logic.

### Binary Patches
Optional client patches for development: custom login screens, extended Lua APIs, custom packet support via DLL injection, and more. See [docs/client-patcher.md](docs/client-patcher.md).

### Mod Bundling
Bundle changes into "mods": Collect DBC, database, LuaXML, and server-side scripts into mods that can be built and distributed easily. Each mod is self-contained with its own migrations, scripts, and assets.

### Local Mod Development
Quickly build and test mods on your local environment. Thorium handles distribution and application of changes in a mod—apply migrations, export DBCs, package MPQs, and sync to your server and client with a single command.

### Distribute Mods
Bundle your mod for sharing: MPQs, SQL files, and scripts packaged together with install instructions.

## Install

See [docs/install.md](docs/install.md) for pre-built binaries or building from source.

## Quick Start

Install an existing mod from GitHub:

```bash
thorium init # Initialize workspace
vim config.json # Configure paths and database connections
thorium get https://github.com/suprsokr/thorium-custom-packets # Install a mod
thorium build # Build the mod
```

Or create a new item mod from scratch:

```bash
thorium init # Initialize workspace
vim config.json # Configure paths and database connections
thorium init db # One-time setup to enable DBC editing via SQL, see docs/dbc.md
thorium create-mod my-custom-item
thorium create-migration --mod my-custom-item --db dbc add_legendary_sword # Edit DBCs via SQL
vim mods/my-custom-item/dbc_sql/*_add_legendary_sword.sql
# DELETE FROM `Item` WHERE `id` = 90001;
# INSERT INTO `Item` (`id`, `class`, `subclass`, `sound_override_subclass`, `material`, `display_id`, `inventory_type`, `sheath`) 
# VALUES (90001, 2, 7, -1, 1, 32254, 17, 1);
thorium create-migration --mod my-custom-item --db world add_legendary_sword_stats # Edit your server to know about your custom item.
vim mods/my-custom-item/world_sql/*_add_legendary_sword_stats.sql
# DELETE FROM `item_template` WHERE `entry` = 90001;
# INSERT INTO `item_template` (
#     `entry`, `class`, `subclass`, `SoundOverrideSubclass`, `name`, `displayid`,
#     `Quality`, `InventoryType`, `ItemLevel`, `RequiredLevel`,
#     `StatsCount`, `stat_type1`, `stat_value1`,
#     `dmg_min1`, `dmg_max1`, `dmg_type1`, `delay`,
#     `bonding`, `description`, `Material`, `sheath`, `MaxDurability`,
#     `RequiredDisenchantSkill`, `DisenchantID`
# ) VALUES (
#     90001, 2, 7, -1, 'Legendary Sword of Power', 32254,
#     5, 17, 80, 70,
#     1, 7, 50,
#     200, 300, 0, 3500,
#     1, 'A legendary weapon of immense power.', 1, 1, 120,
#     -1, 60
# );

# Easily undo changes you don't want later with rollback files.
vim mods/my-custom-item/dbc_sql/*_add_legendary_sword.rollback.sql
# DELETE FROM `Item` WHERE `id` = 90001;
vim mods/my-custom-item/world_sql/*_add_legendary_sword_stats.rollback.sql
# DELETE FROM `item_template` WHERE `entry` = 90001;

# Build everything: apply migrations, package DBCs into MPQs, export DBCs to your WoTLK client.
thorium build
```

Test your mod:
1. Launch your Trinty Core 335 server
2. Launch your WoTLK client. 
3. `.additem 90001` in game from a GM account.

## Ready to distribute your mod to other users?

Once your mod is tested and ready, create a distributable package for server admins:

```bash
# Build everything (applies migrations, exports DBCs, creates MPQs)
thorium build

# Create distribution package (includes client MPQs + server SQL)
thorium dist --output ./releases/my-mod-v1.0.0.zip
```

The generated zip contains:
- **Client MPQs**: Copy to `Data/patch-T.MPQ` and `Data/<locale>/patch-<locale>-T.MPQ`
- **Server SQL**: Migration files to apply to the world database
- **README.txt**: Installation instructions for recipients

Share the zip with server admins. For players, you can `thorium dist --client-only` to skip the server-side edits. See [docs/distribution.md](docs/distribution.md) for detailed installation instructions.