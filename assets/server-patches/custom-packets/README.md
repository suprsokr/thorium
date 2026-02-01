# Custom Packets Server Patch

This patch adds custom packet support to stock TrinityCore 3.3.5a.

## What It Does

- Registers opcode `0x51F` (`CMSG_CUSTOM_PACKET`) for client→server custom packets
- Defines opcode `0x102` (`SMSG_CUSTOM_PACKET`) for server→client custom packets  
- Adds `OnCustomPacketReceive` hook to `ServerScript` for handling custom packets
- Adds `Player::SendCustomPacket(opcode, data)` helper for easy responses
- Parses the ClientExtensions.dll packet header (fragmentId, totalFrags, innerOpcode)

## Requirements

- TrinityCore 3.3.5a (stock or compatible fork)
- Client patched with `thorium patch` (includes ClientExtensions.dll)

## Installation

```bash
# Navigate to your TrinityCore source directory
cd /path/to/TrinityCore

# Apply the patch
git apply /path/to/custom-packets.patch

# Rebuild TrinityCore
cd build
make -j$(nproc)

# Restart your server
```

## Reverting

```bash
cd /path/to/TrinityCore
git apply -R /path/to/custom-packets.patch
```

## Usage Example

Create a script that handles custom packets:

```cpp
#include "ScriptMgr.h"
#include "Player.h"
#include "WorldPacket.h"

// Your custom opcodes (must match client addon)
enum MyOpcodes
{
    CYCM_PING = 1001,  // Client -> Server
    CYSM_PONG = 1002   // Server -> Client
};

class MyCustomPacketHandler : public ServerScript
{
public:
    MyCustomPacketHandler() : ServerScript("MyCustomPacketHandler") { }

    void OnCustomPacketReceive(Player* player, uint16 opcode, WorldPacket& packet) override
    {
        if (opcode != CYCM_PING)
            return;

        // Read data from packet
        uint32 timestamp;
        std::string message;
        packet >> timestamp >> message;

        TC_LOG_INFO("custom", "Player {} sent: {} (ts: {})", 
                    player->GetName(), message, timestamp);

        // Build response payload (just your data, no header needed)
        WorldPacket response;
        response << timestamp;
        response << std::string("pong from server!");

        // SendCustomPacket handles the transport header automatically
        player->SendCustomPacket(CYSM_PONG, &response);
    }
};

void AddSC_my_custom_packets()
{
    new MyCustomPacketHandler();
}
```

## Packet Format

Custom packets use a 6-byte header:

| Field | Size | Description |
|-------|------|-------------|
| fragmentId | 2 bytes | Fragment index (0 for single packets) |
| totalFrags | 2 bytes | Total fragments (1 for single packets) |
| innerOpcode | 2 bytes | Your custom opcode (0-65535) |
| payload | varies | Your packet data |

**Note:** This minimal patch only supports single-fragment packets. For packets larger than ~30KB, you'll need to implement reassembly or extend the patch.

## Opcodes

| Direction | Opcode | Constant |
|-----------|--------|----------|
| Client → Server | 0x51F | CMSG_CUSTOM_PACKET |
| Server → Client | 0x102 | SMSG_CUSTOM_PACKET |

## Credits

- Based on [TSWoW](https://github.com/tswow/tswow) custom packet implementation
- ClientExtensions.dll from https://github.com/suprsokr/wotlk-custom-packets
