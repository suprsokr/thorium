# Configuration

The `config.json` file is the central configuration for your Thorium workspace. It defines paths, database connections, and output settings.

## Location

`config.json` is created in your workspace root when you run `thorium init`. See [Initialization](init.md) for workspace setup.

## Environment Variables

Thorium supports environment variables in two ways:

### Variable Expansion

All paths in `config.json` support environment variable expansion:

- `${VAR}` - Uses the environment variable `VAR`, or empty string if not set
- `${VAR:-default}` - Uses the environment variable `VAR`, or `default` if not set

This allows you to use the same config across different environments. See [Installation](install.md#configuration) for examples.

### Fallback Values

Some configuration fields will automatically use environment variables as fallbacks when the field is empty or not set in `config.json`:

- `databases.dbc.host` and `databases.dbc.port` → `MYSQL_HOST` and `MYSQL_PORT`
- `databases.dbc_source.host` and `databases.dbc_source.port` → `MYSQL_HOST` and `MYSQL_PORT`
- `databases.world.host` and `databases.world.port` → `MYSQL_HOST` and `MYSQL_PORT`
- `trinitycore.source_path` → `TC_SOURCE_PATH`
- `server.dbc_path` → `TC_SERVER_PATH`

**Note:** `wotlk.path` does not use environment variable fallbacks. You must either set it explicitly in `config.json` or use `${WOTLK_PATH}` expansion syntax.

## Configuration Structure

```json
{
  "wotlk": { ... },
  "databases": { ... },
  "server": { ... },
  "trinitycore": { ... },
  "output": { ... }
}
```

## WoW Client Configuration

### `wotlk.path`

**Type:** `string`  
**Default:** `${WOTLK_PATH:-/wotlk}`

Path to your WoW 3.3.5a (12340) client directory. This is used for:

- Binary edits to `Wow.exe` - See [Binary Edits](binary-edits.md)
- Copying assets to client - See [Assets](assets.md)
- Extracting DBC and LuaXML files - See [DBC Workflow](dbc.md) and [LuaXML](luaxml.md)
- Copying MPQ files during build - See [Build](build.md)

**Example:**
```json
{
  "wotlk": {
    "path": "/path/to/wow-folder"
  }
}
```

### `wotlk.locale`

**Type:** `string`  
**Default:** `"enUS"`

Client locale code. Used for:

- Locale-specific LuaXML MPQ naming (e.g., `patch-enUS-T.MPQ`)
- Locale-specific file paths

**Supported locales:** `enUS`, `enGB`, `deDE`, `frFR`, `esES`, `ruRU`, `koKR`, `zhCN`, `zhTW`

**Example:**
```json
{
  "wotlk": {
    "locale": "enUS"
  }
}
```

## Database Configuration

Database connections use the same structure for all database types:

```json
{
  "user": "trinity",
  "password": "trinity",
  "host": "127.0.0.1",
  "port": "3306",
  "name": "database_name"
}
```

### `databases.dbc`

**Type:** `object`  
**Default:** See below

Development database for DBC modifications. SQL migrations in `mods/<mod>/dbc_sql/` are applied to this database. See [DBC Workflow](dbc.md) for details.

**Defaults:**
- `user`: `"trinity"`
- `password`: `"trinity"`
- `host`: `${MYSQL_HOST:-127.0.0.1}`
- `port`: `${MYSQL_PORT:-3306}`
- `name`: `"dbc"`

**Example:**
```json
{
  "databases": {
    "dbc": {
      "user": "trinity",
      "password": "trinity",
      "host": "127.0.0.1",
      "port": "3306",
      "name": "dbc"
    }
  }
}
```

### `databases.dbc_source`

**Type:** `object`  
**Default:** See below

Baseline database containing pristine DBC data. This database is never modified by migrations. Used as a reference to determine which DBCs have been modified. See [DBC Workflow](dbc.md#why-two-databases) for why two databases are needed for player client [Distribution](dist.md).

**Defaults:**
- `user`: `"trinity"`
- `password`: `"trinity"`
- `host`: `${MYSQL_HOST:-127.0.0.1}`
- `port`: `${MYSQL_PORT:-3306}`
- `name`: `"dbc_source"`

**Example:**
```json
{
  "databases": {
    "dbc_source": {
      "user": "trinity",
      "password": "trinity",
      "host": "127.0.0.1",
      "port": "3306",
      "name": "dbc_source"
    }
  }
}
```

### `databases.world`

**Type:** `object`  
**Default:** See below

TrinityCore world database. SQL migrations in `mods/<mod>/world_sql/` are applied to this database. See [SQL Migrations](sql-migrations.md) for details.

**Note:** This database is created and managed by TrinityCore. Thorium only applies migrations to it.

**Defaults:**
- `user`: `"trinity"`
- `password`: `"trinity"`
- `host`: `${MYSQL_HOST:-127.0.0.1}`
- `port`: `${MYSQL_PORT:-3306}`
- `name`: `"world"`

**Example:**
```json
{
  "databases": {
    "world": {
      "user": "trinity",
      "password": "trinity",
      "host": "127.0.0.1",
      "port": "3306",
      "name": "world"
    }
  }
}
```

## Server Configuration

### `server.dbc_path`

**Type:** `string`  
**Default:** `${TC_SERVER_PATH:-/home/peacebloom/server}/bin/dbc`

Path where TrinityCore expects DBC files. During build, modified DBCs are copied here so the server can use them. See [Build](build.md#step-8-package-and-distribute) for details.

**Example:**
```json
{
  "server": {
    "dbc_path": "/path/to/trinitycore/bin/dbc"
  }
}
```

## TrinityCore Configuration

### `trinitycore.source_path`

**Type:** `string`  
**Default:** `${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}`

Path to TrinityCore source code root directory. Used for:

- Applying server patches - See [Server Patches](server-patches.md)
- Deploying C++ scripts - See [Scripts](scripts.md)

**Example:**
```json
{
  "trinitycore": {
    "source_path": "/path/to/TrinityCore"
  }
}
```

### `trinitycore.scripts_path`

**Type:** `string`  
**Default:** `${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}/src/server/scripts/Custom`

Path to TrinityCore's Custom scripts directory. Scripts from `mods/<mod>/scripts/` are deployed here during build. See [Scripts](scripts.md) for details.

**Example:**
```json
{
  "trinitycore": {
    "scripts_path": "/path/to/TrinityCore/src/server/scripts/Custom"
  }
}
```

**Note:** If `source_path` is set, this defaults to `{source_path}/src/server/scripts/Custom`.

## Output Configuration

### `output.dbc_mpq`

**Type:** `string`  
**Default:** `"patch-T.MPQ"`

Filename for the DBC MPQ archive. This file contains modified DBCs and is placed in the client's `Data/` directory during build. See [Build](build.md#step-8-package-and-distribute) and [Distribution](distribution.md) for details.

**Example:**
```json
{
  "output": {
    "dbc_mpq": "patch-T.MPQ"
  }
}
```

### `output.luaxml_mpq`

**Type:** `string`  
**Default:** `"patch-{locale}-T.MPQ"`

Filename template for the LuaXML MPQ archive. The `{locale}` placeholder is replaced with the client locale (e.g., `enUS`). This file contains interface modifications and is placed in the client's `Data/<locale>/` directory during build. See [Build](build.md#step-8-package-and-distribute) and [LuaXML](luaxml.md) for details.

**Example:**
```json
{
  "output": {
    "luaxml_mpq": "patch-{locale}-T.MPQ"
  }
}
```

**Note:** The `{locale}` placeholder is automatically replaced with the value from `wotlk.locale`.

## Complete Example

Here's a complete `config.json` example with all options:

```json
{
  "wotlk": {
    "path": "${WOTLK_PATH:-/wotlk}",
    "locale": "enUS"
  },
  "databases": {
    "dbc": {
      "user": "trinity",
      "password": "trinity",
      "host": "${MYSQL_HOST:-127.0.0.1}",
      "port": "${MYSQL_PORT:-3306}",
      "name": "dbc"
    },
    "dbc_source": {
      "user": "trinity",
      "password": "trinity",
      "host": "${MYSQL_HOST:-127.0.0.1}",
      "port": "${MYSQL_PORT:-3306}",
      "name": "dbc_source"
    },
    "world": {
      "user": "trinity",
      "password": "trinity",
      "host": "${MYSQL_HOST:-127.0.0.1}",
      "port": "${MYSQL_PORT:-3306}",
      "name": "world"
    }
  },
  "server": {
    "dbc_path": "${TC_SERVER_PATH:-/home/peacebloom/server}/bin/dbc"
  },
  "trinitycore": {
    "source_path": "${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}",
    "scripts_path": "${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}/src/server/scripts/Custom"
  },
  "output": {
    "dbc_mpq": "patch-T.MPQ",
    "luaxml_mpq": "patch-{locale}-T.MPQ"
  }
}
```