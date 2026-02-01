# Thorium Get - Installing Mods from GitHub

The `thorium get` command allows you to easily install Thorium mods from GitHub repositories into your workspace.

## Usage

```bash
thorium get <github-url> [--name <custom-name>] [--update]
```

## Flags

- `--name <custom-name>` - Install the mod with a custom name instead of auto-detecting from repository
- `--update` - Update/overwrite an existing mod if it already exists

## Requirements

- Must be run from within a Thorium workspace (a directory containing `mods/config.json`)
- Git must be installed on your system

## Examples

### Install a mod from GitHub

```bash
thorium get https://github.com/suprsokr/thorium-custom-packets
```

This will:
1. Clone the repository to a temporary location
2. Validate it's a valid Thorium mod structure
3. Copy the mod files to your workspace's `mods/` directory
4. Clean up the temporary clone

### Mod naming

The command uses the repository name as the mod name:

- `https://github.com/user/thorium-custom-packets` → installed as `thorium-custom-packets`
- `https://github.com/user/my-awesome-mod` → installed as `my-awesome-mod`

The repository name is used exactly as-is for the mod directory name.

### Custom mod name

You can override the automatic naming with the `--name` flag:

```bash
thorium get https://github.com/user/some-repo --name my-custom-name
```

This installs the mod as `my-custom-name` instead of `some-repo`.

### Updating existing mods

To update an existing mod to the latest version from GitHub:

```bash
thorium get https://github.com/suprsokr/thorium-custom-packets --update
```

This will:
1. Remove the existing mod directory
2. Clone the latest version from GitHub
3. Install the updated mod

**Warning:** This will overwrite any local changes you've made to the mod files.

## Validation

The command validates that the repository contains a valid Thorium mod by checking for at least one of these directories:

- `scripts/` - TrinityCore C++ scripts
- `server-patches/` - Server source code patches
- `binary-edits/` - Client binary patches
- `assets/` - Files to copy to the client
- `luaxml/` - Client Lua/XML modifications
- `dbc_sql/` - DBC database migrations
- `world_sql/` - World database migrations

## After Installation

Once installed, the mod is ready to use:

```bash
# Review the mod's documentation
cat mods/custom-packets/README.md

# Build the mod
thorium build

# Or build just this mod
thorium build --mod custom-packets
```

## Error Handling

### Mod already exists

If a mod with the same name already exists, the command will fail with an error:

```
Error: mod 'thorium-custom-packets' already exists at: mods/thorium-custom-packets

Use --update to overwrite, or --name to install with a different name
```

You have three options:

1. **Update the existing mod:**
   ```bash
   thorium get https://github.com/suprsokr/thorium-custom-packets --update
   ```

2. **Install with a different name:**
   ```bash
   thorium get https://github.com/suprsokr/thorium-custom-packets --name custom-packets-v2
   ```

3. **Manually remove and reinstall:**
   ```bash
   rm -rf mods/thorium-custom-packets
   thorium get https://github.com/suprsokr/thorium-custom-packets
   ```

### Not in a Thorium workspace

If you run `thorium get` outside a Thorium workspace:

```
Error loading config: open config.json: no such file or directory
Run 'thorium init' to create a workspace.
```

Initialize a workspace first:

```bash
thorium init
thorium get https://github.com/suprsokr/thorium-custom-packets
```

### Invalid repository

If the repository doesn't contain valid mod structure:

```
Error: repository does not appear to be a valid Thorium mod (missing expected directories)
```

## Publishing Mods

To make your mod installable via `thorium get`:

1. Create a GitHub repository
2. Push your mod directory contents to the repository root
3. Share the repository URL with users

Example repository structure:

```
your-repo/
├── README.md
├── scripts/
│   └── my_script.cpp
├── server-patches/
│   └── my-patch.patch
├── luaxml/
│   └── Interface/AddOns/...
└── assets/
    └── config.json
```

Users can then install it with:

```bash
thorium get https://github.com/yourusername/your-repo
```

## See Also

- [Installation Guide](install.md)
- [Creating Mods](mods.md)
- [Distribution](distribution.md)
