# Binary Edits

Binary edits allow mods to patch the WoW client executable (`Wow.exe`) with byte-level modifications. This is used for features like DLL injection, memory patches, and client behavior modifications.

## Overview

- Binary edits are JSON files in `mods/<mod>/binary-edits/`
- Applied automatically during `thorium build`
- Tracked in `shared/binary_edits_applied.json` (only applied once)
- A backup of the original `Wow.exe` is saved as `Wow.exe.clean`

## File Format

Create JSON files in your mod's `binary-edits/` folder:

```
mods/my-mod/binary-edits/
├── load-my-dll.json
└── disable-feature.json
```

### Structure

```json
{
  "patches": [
    {
      "address": "0x28e19c",
      "bytes": ["0xE9", "0x19", "0x57", "0x0E", "0x00", "0x90"]
    },
    {
      "address": "0x3738b8",
      "bytes": ["0xEB", "0x26", "0xFF", "0x15", "0x48", "0xF2", "0x9D", "0x00"]
    }
  ]
}
```

### Fields

| Field | Description |
|-------|-------------|
| `patches` | Array of patch objects |
| `patches[].address` | Hex address in the executable (e.g., `"0x28e19c"` or `"28e19c"`) |
| `patches[].bytes` | Array of hex byte values to write (e.g., `["0xE9", "0x19"]`) |

## How It Works

1. During `thorium build`, each `.json` file in `binary-edits/` is processed
2. The backup file `Wow.exe.clean` is verified against the expected MD5 hash for clean WoW 3.3.5a (12340)
3. The patch ID is `<mod-name>/<filename>` (e.g., `custom-packets/load-clientextensions.json`)
4. If the patch ID is not in `shared/binary_edits_applied.json`, the patch is applied
5. Bytes are written to `Wow.exe` at the specified addresses
6. The patch ID is recorded in the tracker

### MD5 Verification

Thorium automatically verifies that `Wow.exe.clean` matches the expected MD5 hash for a clean WoW 3.3.5a (12340) client:

- **Expected MD5**: `45892bdedd0ad70aed4ccd22d9fb5984`
- If the backup matches, you'll see: `✓ Clean WoW 3.3.5a client verified`
- If it doesn't match, you'll see a warning but patching will continue

This ensures you're starting from a known-good client, which is important because:
- Patches are designed for specific byte offsets in the clean client
- Using a modified or different version may cause patches to fail or corrupt the executable
- The warning helps you diagnose issues if patches don't work as expected

## Tracking

Applied edits are tracked in `shared/binary_edits_applied.json`:

```json
{
  "applied": [
    {
      "name": "custom-packets/load-clientextensions.json",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

## Reapplying Edits

To reapply only binary edits (e.g., after restoring from backup):

```bash
thorium build --force-binary-edits
```

Or to force reapply everything (binary edits, server patches, assets, scripts):

```bash
thorium build --force
```

## Backup and Restore

When binary edits are first applied, a backup is created:

```
Wow.exe        ← Patched executable
Wow.exe.clean  ← Original backup (verified against MD5)
```

The backup file should be your original, unmodified WoW 3.3.5a (12340) client. If you need to obtain a clean client:

1. Download the official WoW 3.3.5a (build 12340) client
2. Verify its MD5: `md5sum Wow.exe` (Linux/Mac) or `Get-FileHash Wow.exe -Algorithm MD5` (Windows)
3. Expected MD5: `45892bdedd0ad70aed4ccd22d9fb5984`
4. Copy it to your client directory as `Wow.exe.clean`

To restore the original:

```bash
cp Wow.exe.clean Wow.exe
```

Then delete the relevant entries from `shared/binary_edits_applied.json` if you want `thorium build` to reapply them.

## Tips

- **Test patches** on a copy of `Wow.exe` first
- **Document your patches** - include comments in a README explaining what each patch does
- **Verify your client** - The clean WoW 3.3.5a (12340) client has MD5 `45892bdedd0ad70aed4ccd22d9fb5984`
  - On Linux/Mac: `md5sum Wow.exe`
  - On Windows: `Get-FileHash Wow.exe -Algorithm MD5`
- **Keep a clean backup** - Store `Wow.exe.clean_original` as an untouched backup if you need to restore to pristine state
- **Use descriptive names** - Name your binary edit files clearly (e.g., `load-custom-dll.json`, `enable-feature-x.json`)

## See Also

- [mods.md](mods.md) - Mod structure overview
- [assets.md](assets.md) - Copying files like DLLs to client
- [custom-packets.md](custom-packets.md) - Example mod using binary edits
