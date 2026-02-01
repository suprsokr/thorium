# Server Patches

Server patches allow mods to modify TrinityCore source code using Git patch files. This is used for features that require core changes, like registering new opcodes or adding script hooks.

## Overview

- Server patches are `.patch` files in `mods/<mod>/server-patches/`
- Applied automatically during `thorium build` (if `trinitycore.source_path` is configured)
- Tracked in `shared/server_patches_applied.json` (only applied once)
- After patches are applied, you must rebuild TrinityCore

## File Format

Place standard Git patch files in your mod's `server-patches/` folder:

```
mods/my-mod/server-patches/
├── custom-packets.patch
└── extended-api.patch
```

## Creating Patches

### From Uncommitted Changes

```bash
# Make changes in your TrinityCore source
cd /path/to/TrinityCore
# ... edit files ...

# Create a patch file
git diff > /path/to/workspace/mods/my-mod/server-patches/my-changes.patch
```

### From Staged Changes

```bash
git diff --cached > my-changes.patch
```

### From Commits

```bash
# Single commit
git format-patch -1 HEAD

# Range of commits
git format-patch origin/master..HEAD
```

## How It Works

1. During `thorium build`, each `.patch` file in `server-patches/` is discovered
2. The patch ID is `<mod-name>/<filename>` (e.g., `custom-packets/custom-packets.patch`)
3. If the patch ID is not in `shared/server_patches_applied.json`:
   - `git apply --check` verifies the patch applies cleanly
   - `git apply` applies the patch to TrinityCore source
   - The patch ID is recorded in the tracker
4. If a patch doesn't apply cleanly, it's skipped with a warning

## Tracking

Applied patches are tracked in `shared/server_patches_applied.json`:

```json
{
  "applied": [
    {
      "name": "custom-packets/custom-packets.patch",
      "version": "1.0.0",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

## Reapplying Patches

To reapply only server patches (e.g., after reverting):

```bash
thorium build --force-server-patches
```

Or to force reapply everything (binary edits, server patches, assets, scripts):

```bash
thorium build --force
```

## Reverting Patches

To revert a patch manually:

```bash
cd /path/to/TrinityCore
git apply -R /path/to/mods/my-mod/server-patches/my-patch.patch
```

Then remove the entry from `shared/server_patches_applied.json`.

## Rebuilding TrinityCore

After patches are applied, you must rebuild TrinityCore:

```bash
cd /path/to/TrinityCore/build
make -j$(nproc)
```

Then restart your server.

## Configuration

Set the TrinityCore source path in `config.json`:

```json
{
  "trinitycore": {
    "source_path": "/path/to/TrinityCore"
  }
}
```

Or use the environment variable:

```bash
export TC_SOURCE_PATH=/path/to/TrinityCore
```

## Example: Adding a Script Hook

This patch adds a custom packet handler hook to TrinityCore:

```diff
diff --git a/src/server/game/Scripting/ScriptMgr.h b/src/server/game/Scripting/ScriptMgr.h
--- a/src/server/game/Scripting/ScriptMgr.h
+++ b/src/server/game/Scripting/ScriptMgr.h
@@ -499,6 +499,9 @@ class TC_GAME_API ServerScript : public ScriptObject
 
         // Called when a packet is received.
         virtual void OnPacketReceive(WorldSession* /*session*/, WorldPacket& /*packet*/) { }
+
+        // Called when a custom packet is received from client addon
+        virtual void OnCustomPacketReceive(Player* /*player*/, uint16 /*opcode*/, WorldPacket& /*packet*/) { }
 };
```

## Tips

- **Test patches manually first** with `git apply --check`
- **Keep patches minimal** - only include necessary changes
- **Document changes** - add comments or a README explaining what the patch does
- **Consider upstream** - if your patch is generally useful, consider submitting it upstream

## Troubleshooting

### Patch doesn't apply cleanly

The most common causes:
1. **Already applied** - Check if it's in `server_patches_applied.json`
2. **TrinityCore version mismatch** - The patch was created for a different version
3. **Conflicting changes** - Another mod or manual edit conflicts

To debug:

```bash
cd /path/to/TrinityCore
git apply --check /path/to/patch.patch
```

### Patch applied but feature doesn't work

Remember to:
1. Rebuild TrinityCore after applying patches
2. Restart the worldserver
3. Check server logs for errors

## See Also

- [mods.md](mods.md) - Mod structure overview
- [scripts.md](scripts.md) - TrinityCore C++ scripts
- [custom-packets.md](custom-packets.md) - Example mod using server patches
