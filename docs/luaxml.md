# LuaXML (Interface Files)

LuaXML refers to the Lua scripts and XML layout files that make up World of Warcraft's user interface. Thorium can extract, modify, and package these files for UI customization.

## Overview

The WoW client UI is built from:

- **XML files** - Define frame layouts, templates, and widget structure
- **Lua files** - Contain logic, event handlers, and game interaction code
- **TOC files** - Table of contents that list which files to load

These files are stored in the client's MPQ archives under the `Interface/` directory.

## Directory Structure

```
mods/
├── shared/
│   └── luaxml/
│       └── luaxml_source/           # Baseline files (extracted from client)
│           └── Interface/
│               ├── FrameXML/        # Core UI frames
│               ├── GlueXML/         # Login screen UI
│               └── AddOns/
│                   └── CustomPackets/  # Auto-generated (thorium init)
└── mods/
    └── my-mod/
        └── luaxml/                  # Your modifications (only changed files)
            └── Interface/
                └── FrameXML/
                    └── ChatFrame.lua
```

## Workflow

### 1. Extract Baseline Files

Extract interface files from the WoW client to use as reference:

```bash
# Extract all interface files (~16K files, ~500MB)
thorium extract --luaxml

# Or extract only what you need (recommended)
thorium extract --luaxml --filter Interface/FrameXML
thorium extract --luaxml --filter Interface/AddOns/Blizzard_CombatText
```

### 2. Copy Files to Your Mod

Copy specific files you want to modify into your mod:

```bash
# Copy a single file
thorium extract --mod my-mod --dest Interface/FrameXML/ChatFrame.lua

# Copy an entire addon folder
thorium extract --mod my-mod --dest Interface/AddOns/Blizzard_AchievementUI
```

### 3. Edit the Files

Edit the copied files in your mod's `luaxml/` directory. Only include files you're actually modifying.

### 4. Build

```bash
thorium build --mod my-mod
```

Thorium compares your mod's files against the baseline and packages only the differences into `patch-enUS-T.MPQ`.

## Use Cases

### Modifying Chat Behavior

```bash
thorium extract --mod my-mod --dest Interface/FrameXML/ChatFrame.lua
# Edit mods/my-mod/luaxml/Interface/FrameXML/ChatFrame.lua
```

### Custom Login Screen

With the `allow-custom-gluexml` client patch applied, you can modify login screen files:

```bash
thorium extract --mod my-mod --dest Interface/GlueXML/AccountLogin.lua
```

### Custom Fonts and Textures

Add or replace interface assets in your mod:

```
mods/my-mod/luaxml/Interface/
├── Fonts/
│   └── MyCustomFont.ttf
└── Icons/
    └── MyCustomIcon.blp
```

## Building

When you run `thorium build`, LuaXML files are packaged into a locale-specific MPQ:

```
WoW 3.3.5a/
└── Data/
    └── enUS/
        └── patch-enUS-T.MPQ    ← Your LuaXML changes
```

The client's patch load order ensures your customizations override the defaults.

## Addons

You can create custom addons that are packaged into the LuaXML MPQ:

```bash
thorium create-addon --mod my-mod MyAddon
```

This creates:
```
mods/my-mod/luaxml/Interface/AddOns/MyAddon/
├── MyAddon.toc    # Addon metadata
└── main.lua       # Main addon code
```

### CustomPackets Addon

When you run `thorium init`, a `CustomPackets` addon is automatically created in `shared/luaxml/luaxml_source/Interface/AddOns/CustomPackets/`. This provides a Lua API for custom client-server communication.

Your addons can depend on it:

```toc
## Interface: 30300
## Title: My Addon
## Dependencies: CustomPackets

main.lua
```

Then use the API:

```lua
-- Send custom packet
local packet = CreateCustomPacket(1001, 0)
packet:WriteUInt32(12345)
packet:WriteString("Hello")
packet:Send()

-- Receive custom packet
OnCustomPacket(1002, function(reader)
    local value = reader:ReadUInt32()
    print("Got: " .. value)
end)
```

See [custom-packets.md](custom-packets.md) for more details.

## Notes

- **Only include modified files** in your mod's `luaxml/` folder
- GlueXML modifications require the `allow-custom-gluexml` client patch
- Test UI changes carefully - errors can prevent the client from loading
- Hidden files (starting with `.`) are automatically excluded from packaging
