package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thorium-cli/internal/config"
)

// CreateScript creates a new TrinityCore script file in a mod
func CreateScript(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("create-script", flag.ExitOnError)
	modName := fs.String("mod", "", "Mod name (required)")
	scriptType := fs.String("type", "spell", "Script type: spell, aura, creature, server, packet (default: spell)")
	fs.Parse(args)

	remaining := fs.Args()
	if len(remaining) == 0 {
		return fmt.Errorf("script name required\nUsage: thorium create-script --mod <mod> --type <spell|aura|creature|server|packet> <name>")
	}

	scriptName := strings.Join(remaining, "_")
	scriptName = sanitizeName(scriptName)

	// Validate required flags
	if *modName == "" {
		return fmt.Errorf("--mod flag is required")
	}

	// Validate script type
	validTypes := []string{"spell", "aura", "creature", "server", "packet"}
	isValid := false
	for _, t := range validTypes {
		if *scriptType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("--type must be one of: %v", validTypes)
	}

	// Check mod exists
	modPath := filepath.Join(cfg.GetModsPath(), *modName)
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return fmt.Errorf("mod not found: %s\nRun 'thorium create-mod %s' first", *modName, *modName)
	}

	// Ensure scripts directory exists
	scriptsDir := filepath.Join(modPath, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return fmt.Errorf("create scripts directory: %w", err)
	}

	// Generate filename
	fileName := fmt.Sprintf("%s_%s.cpp", *scriptType, scriptName)
	filePath := filepath.Join(scriptsDir, fileName)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("script already exists: %s", filePath)
	}

	// Generate script content
	content := generateScriptTemplate(*scriptType, scriptName)

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write script file: %w", err)
	}

	fmt.Printf("Created script in %s:\n", *modName)
	fmt.Printf("  %s\n", fileName)
	fmt.Println()
	fmt.Printf("Edit your script:\n")
	fmt.Printf("  %s\n", filePath)
	fmt.Println()
	fmt.Printf("After editing, run 'thorium build' to deploy to TrinityCore source.\n")

	return nil
}

// generateScriptTemplate generates the appropriate template based on script type
func generateScriptTemplate(scriptType, name string) string {
	switch scriptType {
	case "spell":
		return generateSpellScriptTemplate(name)
	case "aura":
		return generateAuraScriptTemplate(name)
	case "creature":
		return generateCreatureScriptTemplate(name)
	case "server":
		return generateServerScriptTemplate(name)
	case "packet":
		return generatePacketScriptTemplate(name)
	default:
		return ""
	}
}

// generateSpellScriptTemplate generates a SpellScript template
func generateSpellScriptTemplate(name string) string {
	className := fmt.Sprintf("spell_%s", name)
	addSCFunc := fmt.Sprintf("AddSC_spell_%s", name)

	return fmt.Sprintf(`// %s - SpellScript

#include "ScriptMgr.h"
#include "SpellScript.h"

// TODO: Set the spell ID this script handles
class %s : public SpellScript
{
    PrepareSpellScript(%s);

    void Register() override
    {
    }
};

void %s()
{
    RegisterSpellScript(%s);
}
`, name, className, className, addSCFunc, className)
}

// generateAuraScriptTemplate generates an AuraScript template
func generateAuraScriptTemplate(name string) string {
	className := fmt.Sprintf("aura_%s", name)
	addSCFunc := fmt.Sprintf("AddSC_aura_%s", name)

	return fmt.Sprintf(`// %s - AuraScript

#include "ScriptMgr.h"
#include "SpellScript.h"
#include "SpellAuraEffects.h"

// TODO: Set the spell ID whose aura this script handles
class %s : public AuraScript
{
    PrepareAuraScript(%s);

    void Register() override
    {
    }
};

void %s()
{
    RegisterSpellScript(%s);
}
`, name, className, className, addSCFunc, className)
}

// generateCreatureScriptTemplate generates a CreatureScript template
func generateCreatureScriptTemplate(name string) string {
	className := fmt.Sprintf("npc_%s", name)
	addSCFunc := fmt.Sprintf("AddSC_npc_%s", name)

	return fmt.Sprintf(`// %s - CreatureScript

#include "ScriptMgr.h"
#include "Creature.h"
#include "ScriptedCreature.h"

// TODO: Set creature_template.ScriptName = "%s" for the creature entry
class %s : public CreatureScript
{
public:
    %s() : CreatureScript("%s") { }

    struct %sAI : public ScriptedAI
    {
        %sAI(Creature* creature) : ScriptedAI(creature) { }

        void UpdateAI(uint32 /*diff*/) override
        {
        }
    };

    CreatureAI* GetAI(Creature* creature) const override
    {
        return new %sAI(creature);
    }
};

void %s()
{
    new %s();
}
`, name, className, className, className, className, className, className, className, addSCFunc, className)
}

// generateServerScriptTemplate generates a ServerScript template
func generateServerScriptTemplate(name string) string {
	className := fmt.Sprintf("%sServerScript", strings.Title(name))
	addSCFunc := fmt.Sprintf("AddSC_%s_server", name)

	return fmt.Sprintf(`// %s - ServerScript

#include "ScriptMgr.h"

class %s : public ServerScript
{
public:
    %s() : ServerScript("%s") { }
};

void %s()
{
    new %s();
}
`, name, className, className, className, addSCFunc, className)
}

// generatePacketScriptTemplate generates a ServerScript for custom packets
func generatePacketScriptTemplate(name string) string {
	className := fmt.Sprintf("%sPacketScript", strings.Title(name))
	addSCFunc := fmt.Sprintf("AddSC_%s_packet", name)

	return fmt.Sprintf(`// %s - Custom Packet Handler

#include "ScriptMgr.h"
#include "WorldSession.h"
#include "WorldPacket.h"

class %s : public ServerScript
{
public:
    %s() : ServerScript("%s") { }

    void OnPacketReceive(WorldSession* session, WorldPacket& packet) override
    {
    }
};

void %s()
{
    new %s();
}
`, name, className, className, className, addSCFunc, className)
}
