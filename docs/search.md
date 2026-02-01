# Thorium Search - Discovering Mods from the Registry

The `thorium search` command allows you to discover and explore Thorium mods from the official mod registry.

## Overview

The [Thorium Mod Registry](https://github.com/suprsokr/thorium-mod-registry) is a curated collection of publicly available mods hosted on GitHub. The `search` command queries this registry to help you find mods that match your needs, whether you're looking for specific functionality, browsing by category, or exploring what's available.

## Usage

```bash
thorium search [query] [--tag <tag>] [--name <mod-name>] [--tags]
```

## Basic Search

### Search by keyword

Search across mod names, descriptions, authors, and tags:

```bash
# Find mods related to networking
thorium search networking

# Find UI-related mods
thorium search ui

# Find mods by author
thorium search suprsokr

# Show all available mods
thorium search
```

Keywords are matched against:
- Mod name
- Display name
- Description
- Author name
- Tags

The search is case-insensitive and matches partial words.

## Tag-Based Search

### Search by single tag

```bash
thorium search --tag lua-api
thorium search --tag dungeons
thorium search --tag client-only
```

### Search by multiple tags (AND logic)

When you specify multiple tags, only mods that have **all** specified tags are returned:

```bash
# Find mods that are both networking-related AND provide an API
thorium search --tag networking --tag api

# Find client-side UI mods
thorium search --tag ui --tag client-only
```

### Combine keyword and tags

```bash
# Find networking mods with framework tag
thorium search networking --tag framework

# Find UI mods by specific author
thorium search suprsokr --tag ui
```

### List all available tags

See all tags currently used in the registry:

```bash
thorium search --tags
```

## Viewing Mod Details

### Get detailed information about a specific mod

```bash
thorium search --name thorium-custom-packets
```

This displays:
- Full description
- Version and license information
- Repository and homepage URLs
- Complete tag list
- Compatibility information
- Dependencies (if any)
- Statistics (stars, downloads if tracked)
- Installation command

## Search Results

### Summary View

When searching by keyword or tag, you get a summary view:

```
Found 3 mod(s):

┌─ Custom Packets (v1.0.0)
│  Name: thorium-custom-packets
│  Author: suprsokr
│  Client-server packet communication framework with Lua API
│  Tags: networking, lua-api, client-server, framework, api
│  Repository: https://github.com/suprsokr/thorium-custom-packets
└─

┌─ Enhanced UI Pack (v2.1.0)
│  Name: enhanced-ui-pack
│  Author: uimaster
│  Modern UI improvements with quality-of-life features
│  Tags: ui, quality-of-life, client-only, graphics
│  Repository: https://github.com/uimaster/enhanced-ui-pack
└─

To install a mod: thorium get <repository-url>
For details:       thorium search --name <mod-name>
List all tags:     thorium search --tags
```

### Detailed View

When viewing a specific mod with `--name`:

```
╔═══════════════════════════════════════════════════════════════
║ Custom Packets
╠═══════════════════════════════════════════════════════════════
║
║ Name:        thorium-custom-packets
║ Version:     1.0.0
║ Author:      suprsokr
║ License:     MIT
║
║ Description:
║   Client-server packet communication framework with Lua API
║   for creating custom network protocols
║
║ Repository:  https://github.com/suprsokr/thorium-custom-packets
║
║ Tags:        networking, lua-api, client-server, framework, api
║ Compatible:  3.3.5a
║
║ Added:       2026-01-15
║ Updated:     2026-01-31
║
╚═══════════════════════════════════════════════════════════════

To install this mod:
  thorium get https://github.com/suprsokr/thorium-custom-packets
```

## After Finding a Mod

Once you find a mod you want to install:

### 1. View full details (optional)

```bash
thorium search --name mod-name-here
```

### 2. Install the mod

```bash
thorium get https://github.com/author/repository
```

### 3. Check for dependencies

If the mod requires other mods, install dependencies first:

```bash
# The detailed view shows required mods
thorium search --name dependent-mod

# Install each dependency
thorium get https://github.com/author/dependency-mod
```

### 4. Build and test

```bash
thorium build
```

## Registry Information

### How often is it updated?

The registry is updated whenever:
- New mods are submitted
- Existing mods are updated
- Mod information changes

The `search` command fetches the latest version from GitHub each time you run it.

### Can I add my mod?

Yes! See the [registry submission guide](https://github.com/suprsokr/thorium-mod-registry) for instructions on adding your mod to the registry.

## Error Handling

### Registry unavailable

If the registry cannot be fetched (network issues, GitHub down, etc.):

```
Error: fetch registry: failed to fetch registry: Get "https://raw.githubusercontent.com/...": dial tcp: lookup raw.githubusercontent.com: no such host
```

Check your internet connection and try again.

### No results

If no mods match your search:

```
No mods found matching your search criteria.
```

Try:
- Broader keywords
- Different tags
- Checking for typos
- Listing all available tags with `thorium search --tags`

### Mod not found

If you search for a specific mod name that doesn't exist:

```
Mod not found: nonexistent-mod

Search for mods with: thorium search <keyword>
```

Verify the mod name or browse available mods with `thorium search`.

## See Also

- [Installing Mods with `thorium get`](https://github.com/suprsokr/thorium/tree/main/docs/get.md)
- [Creating Your Own Mods](https://github.com/suprsokr/thorium/tree/main/docs/mods.md)
- [Distributing a Mod To Players Guide](https://github.com/suprsokr/thorium/tree/main/docs/distribution.md)
