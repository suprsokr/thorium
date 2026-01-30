# LuaXML (Interface Files)

LuaXML refers to the Lua scripts and XML layout files that make up World of Warcraft's user interface. Thorium can extract and package these files for UI customization.

## Overview

The WoW client UI is built from:

- **XML files** - Define frame layouts, templates, and widget structure
- **Lua files** - Contain logic, event handlers, and game interaction code
- **TOC files** - Table of contents that list which files to load

These files are stored in the client's MPQ archives under the `Interface/` directory.

## Directory Structure

```
mods/
└── shared/
    └── luaxml/              # Extracted interface files
        └── Interface/
            ├── FrameXML/    # Core UI frames
            ├── GlueXML/     # Login screen UI
            └── AddOns/      # Built-in addons
```

## Use Cases

### Custom Login Screen

With the `allow-custom-gluexml` client patch applied, you can modify login screen files:

```
mods/shared/luaxml/Interface/GlueXML/
├── AccountLogin.lua
├── AccountLogin.xml
└── GlueLocalization.lua
```

### UI Modifications

Modify the default UI by editing FrameXML files. Common targets:

- `CharacterFrame.xml` - Character panel layout
- `SpellBookFrame.xml` - Spellbook UI
- `QuestFrame.xml` - Quest log and tracking

### Custom Fonts and Textures

Add or replace interface assets:

```
mods/shared/luaxml/Interface/
├── Fonts/
│   └── MyCustomFont.ttf
└── Icons/
    └── MyCustomIcon.blp
```

## Building

When you run `thorium build`, LuaXML files are packaged into a locale-specific MPQ. The output file follows the pattern `patch-<locale>-T.MPQ` (e.g., `patch-enUS-T.MPQ`).

This file is installed to the client's locale directory:

```
WoW 3.3.5a/
└── Data/
    └── enUS/
        └── patch-enUS-T.MPQ    ← LuaXML patch goes here
```

The client's patch load order ensures your customizations override the defaults.

## Notes

- GlueXML modifications require the `allow-custom-gluexml` client patch
- FrameXML changes are applied via MPQ patch priority
- Test UI changes carefully - errors can prevent the client from loading
