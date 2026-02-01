# Custom Packets

Custom Packets enable bidirectional communication between WoW client addons and TrinityCore server scripts beyond the standard WoW protocol. This allows you to build features like custom UI updates, real-time data sync, and server-driven client behavior.

## Overview

The system consists of four parts:

1. **Client-side Lua API** - `CustomxPackets` addon for sending/receiving packets in addons
2. **Client binary patch** - Injects `ClientExtensions.dll` to handle custom opcodes
3. **Server source patch** - Patches TrinityCore to register opcode 0x51F
4. **Server-side C++ handlers** - Your scripts that process custom packets

**Important:** Custom packets require patches to BOTH the WoW client AND TrinityCore server. Stock TrinityCore does not support custom opcodes out of the box.

## Architecture

```
┌─────────────────────┐                    ┌─────────────────────┐
│   WoW Client        │                    │   TrinityCore       │
│                     │                    │                     │
│  ┌───────────────┐  │    Custom Opcode   │  ┌───────────────┐  │
│  │ Your Addon    │  │ ──────────────────►│  │ Packet Script │  │
│  │               │  │    (0x51F)         │  │               │  │
│  │ CreateCustom- │  │                    │  │ OnCustom-     │  │
│  │ Packet()      │  │                    │  │ Packet()      │  │
│  └───────────────┘  │                    │  └───────────────┘  │
│         ▲           │                    │         │           │
│         │           │    Custom Opcode   │         │           │
│         └───────────│◄───────────────────│─────────┘           │
│                     │    (0x102)         │                     │
│  ┌───────────────┐  │                    │                     │
│  │ CustomPackets │  │                    │                     │
│  │ Addon (API)   │  │                    │                     │
│  └───────────────┘  │                    │                     │
└─────────────────────┘                    └─────────────────────┘
```

### Opcodes

| Direction | Opcode | Description |
|-----------|--------|-------------|
| Client → Server | `0x51F` | Custom packets sent from addons |
| Server → Client | `0x102` | Custom packets sent from server scripts |

### Packet Structure

Each custom packet has a 6-byte header:

| Field | Size | Description |
|-------|------|-------------|
| FragmentID | 2 bytes | Current fragment index (0-based) |
| TotalFrags | 2 bytes | Total number of fragments |
| Opcode | 2 bytes | Your custom opcode (0-65535) |

The header is followed by the payload data. Large packets are automatically fragmented (max ~30KB per fragment).

## Client-Side: Lua API

The `CustomPackets` addon is created automatically during `thorium init` and provides the full Lua API.

### Sending Packets

```lua
-- Create a packet with your custom opcode
local packet = CreateCustomPacket(1001, 0)  -- opcode, size hint (0 = dynamic)

-- Write data
packet:WriteUInt8(255)
packet:WriteInt32(-12345)
packet:WriteFloat(3.14159)
packet:WriteString("Hello Server")          -- null-terminated
packet:WriteLengthString("Length prefixed") -- uint32 length + bytes

-- Send to server
packet:Send()
```

### Receiving Packets

```lua
-- Register handler for custom opcode
OnCustomPacket(1002, function(reader)
    -- Read data in same order it was written
    local flags = reader:ReadUInt8(0)           -- 0 is default if read fails
    local count = reader:ReadInt32(0)
    local multiplier = reader:ReadFloat(1.0)
    local name = reader:ReadString("")          -- null-terminated
    local message = reader:ReadLengthString("") -- length-prefixed
    
    print("Received:", name, message)
end)

-- Unregister handler
OffCustomPacket(1002)
```

### Data Types

| Write Method | Read Method | Size | Range |
|--------------|-------------|------|-------|
| `WriteUInt8(v)` | `ReadUInt8(def)` | 1 byte | 0 to 255 |
| `WriteInt8(v)` | `ReadInt8(def)` | 1 byte | -128 to 127 |
| `WriteUInt16(v)` | `ReadUInt16(def)` | 2 bytes | 0 to 65,535 |
| `WriteInt16(v)` | `ReadInt16(def)` | 2 bytes | -32,768 to 32,767 |
| `WriteUInt32(v)` | `ReadUInt32(def)` | 4 bytes | 0 to 4,294,967,295 |
| `WriteInt32(v)` | `ReadInt32(def)` | 4 bytes | -2.1B to 2.1B |
| `WriteUInt64(v)` | `ReadUInt64(def)` | 8 bytes | 0 to 18.4E |
| `WriteInt64(v)` | `ReadInt64(def)` | 8 bytes | -9.2E to 9.2E |
| `WriteFloat(v)` | `ReadFloat(def)` | 4 bytes | IEEE 754 single |
| `WriteDouble(v)` | `ReadDouble(def)` | 8 bytes | IEEE 754 double |
| `WriteString(v)` | `ReadString(def)` | varies | Null-terminated |
| `WriteLengthString(v)` | `ReadLengthString(def)` | varies | uint32 length + bytes |
| `WriteBytes(t)` | `ReadBytes(n)` | varies | Raw byte array |

**Note:** All read methods take a default value that is returned if the read fails (e.g., end of packet).

### Utility Methods

```lua
-- Writer
packet:Size()              -- Current packet size in bytes

-- Reader
reader:Remaining()         -- Bytes left to read
reader:Skip(count)         -- Skip N bytes
reader:Reset()             -- Reset read position to start
```

## Server-Side: C++ Scripts

Create a packet handler script:

```bash
thorium create-script --mod my-mod --type packet my_protocol
```

This generates a script template in `mods/my-mod/scripts/`.

### Example Handler

```cpp
#include "ScriptMgr.h"
#include "Player.h"
#include "WorldPacket.h"

// Define your opcodes (must match client-side)
enum MyOpcodes
{
    CYCM_MY_REQUEST  = 1001,  // Client -> Server
    CYSM_MY_RESPONSE = 1002   // Server -> Client
};

class MyCustomPacketHandler : public ServerScript
{
public:
    MyCustomPacketHandler() : ServerScript("MyCustomPacketHandler") { }

    // This hook is added by the custom-packets server patch
    void OnCustomPacketReceive(Player* player, uint16 opcode, WorldPacket& packet) override
    {
        if (opcode != CYCM_MY_REQUEST)
            return;

        // Read data in same order client wrote it
        uint8 flags;
        int32 count;
        float multiplier;
        std::string name;

        packet >> flags >> count >> multiplier >> name;

        // Process the data
        TC_LOG_INFO("custom", "Received from {}: flags={}, count={}, name={}",
            player->GetName(), flags, count, name);

        // Build response payload (just the data, no header needed)
        WorldPacket response;
        response << uint32(player->GetGUID().GetCounter());
        response << std::string("Response from server");
        
        // SendCustomPacket handles the transport header automatically
        player->SendCustomPacket(CYSM_MY_RESPONSE, &response);
    }
};

void AddSC_my_custom_packets()
{
    new MyCustomPacketHandler();
}
```

## Setup

### 1. Initialize Workspace

```bash
thorium init
```

This creates the `CustomPackets` addon in `shared/luaxml/luaxml_source/Interface/AddOns/CustomPackets/`.

### 2. Apply Client Patches

```bash
thorium patch
```

This applies the `custom-packets` patch (among others) that:
- Patches `WoW.exe` to load `ClientExtensions.dll`
- Copies `ClientExtensions.dll` to your WoW directory

### 3. Apply Server Patches

```bash
# List available server patches
thorium patch-server --list

# Apply the custom-packets patch to TrinityCore source
thorium patch-server apply custom-packets

# Rebuild TrinityCore
cd /path/to/TrinityCore/build && make -j$(nproc)
```

This patches TrinityCore to:
- Register opcode `0x51F` (`CMSG_CUSTOM_PACKET`)
- Add `OnCustomPacketReceive` hook to `ServerScript`
- Add `Player::SendCustomPacket(opcode, data)` helper for easy responses

**Note:** You must rebuild TrinityCore after applying the server patch.

### 4. Create Your Addon

```bash
thorium create-addon --mod my-mod MyFeatureUI
```

Edit `mods/my-mod/luaxml/Interface/AddOns/MyFeatureUI/MyFeatureUI.toc`:

```toc
## Interface: 30300
## Title: My Feature UI
## Dependencies: CustomPackets

main.lua
```

### 5. Create Server Handler

```bash
thorium create-script --mod my-mod --type packet my_feature_protocol
```

### 6. Build and Test

```bash
# Build client files (packages addons into MPQ)
thorium build

# Rebuild TrinityCore with your script (if you added new scripts)
cd /path/to/TrinityCore/build && make -j$(nproc)

# Restart server and test in-game
```

## See Also

- [client-patcher.md](client-patcher.md) - Binary patches including custom-packets
- [luaxml.md](luaxml.md) - Addon packaging and LuaXML
- [scripts.md](scripts.md) - TrinityCore script creation
- [mods.md](mods.md) - Mod structure overview
