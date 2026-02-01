// Copyright (c) 2025 Thorium

package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"thorium-cli/internal/config"
)

const registryURL = "https://raw.githubusercontent.com/suprsokr/thorium-mod-registry/main/registry.json"

// RegistryMod represents a mod entry in the registry
type RegistryMod struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	Repository  string    `json:"repository"`
	Tags        []string  `json:"tags"`
	Version     string    `json:"version"`
	Requires    []string  `json:"requires"`
	AddedDate   time.Time `json:"added_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

// Registry represents the mod registry
type Registry struct {
	Version     string        `json:"version"`
	LastUpdated time.Time     `json:"last_updated"`
	Mods        []RegistryMod `json:"mods"`
}

// Search searches the Thorium mod registry
func Search(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	tagsFlag := fs.String("tag", "", "Filter by tag (can specify multiple times with comma separation)")
	nameFlag := fs.String("name", "", "Show full details for a specific mod by exact name")
	allTags := fs.Bool("tags", false, "List all available tags")
	fs.Parse(args)

	// Fetch registry
	registry, err := fetchRegistry()
	if err != nil {
		return fmt.Errorf("fetch registry: %w", err)
	}

	// List all tags mode
	if *allTags {
		return listAllTags(registry)
	}

	// Show specific mod details
	if *nameFlag != "" {
		return showModDetails(registry, *nameFlag)
	}

	// Parse search query
	var searchQuery string
	if fs.NArg() > 0 {
		searchQuery = strings.ToLower(strings.Join(fs.Args(), " "))
	}

	// Parse tags filter
	var tags []string
	if *tagsFlag != "" {
		tags = strings.Split(*tagsFlag, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(strings.ToLower(tags[i]))
		}
	}

	// Filter mods
	results := filterMods(registry.Mods, searchQuery, tags)

	// Display results
	if len(results) == 0 {
		fmt.Println("No mods found matching your search criteria.")
		return nil
	}

	// Sort results by name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	fmt.Printf("Found %d mod(s):\n\n", len(results))
	for _, mod := range results {
		printModSummary(mod)
	}

	fmt.Printf("\nTo install a mod: thorium get <repository-url>\n")
	fmt.Printf("For details:       thorium search --name <mod-name>\n")
	fmt.Printf("List all tags:     thorium search --tags\n")

	return nil
}

// fetchRegistry fetches the mod registry from GitHub
func fetchRegistry() (*Registry, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(registryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}

	var registry Registry
	if err := json.Unmarshal(body, &registry); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}

	return &registry, nil
}

// filterMods filters mods based on search query and tags
func filterMods(mods []RegistryMod, query string, tags []string) []RegistryMod {
	var results []RegistryMod

	for _, mod := range mods {
		// Check tags filter first (AND logic - must have all specified tags)
		if len(tags) > 0 {
			hasAllTags := true
			for _, requiredTag := range tags {
				found := false
				for _, modTag := range mod.Tags {
					if strings.ToLower(modTag) == requiredTag {
						found = true
						break
					}
				}
				if !found {
					hasAllTags = false
					break
				}
			}
			if !hasAllTags {
				continue
			}
		}

		// If no query, include all mods that passed tag filter
		if query == "" {
			results = append(results, mod)
			continue
		}

		// Search in name, display name, description, author, and tags
		lowerName := strings.ToLower(mod.Name)
		lowerDisplayName := strings.ToLower(mod.DisplayName)
		lowerDescription := strings.ToLower(mod.Description)
		lowerAuthor := strings.ToLower(mod.Author)

		if strings.Contains(lowerName, query) ||
			strings.Contains(lowerDisplayName, query) ||
			strings.Contains(lowerDescription, query) ||
			strings.Contains(lowerAuthor, query) {
			results = append(results, mod)
			continue
		}

		// Check tags
		for _, tag := range mod.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, mod)
				break
			}
		}
	}

	return results
}

// printModSummary prints a brief summary of a mod
func printModSummary(mod RegistryMod) {
	fmt.Printf("┌─ %s (v%s)\n", mod.DisplayName, mod.Version)
	fmt.Printf("│  Name: %s\n", mod.Name)
	fmt.Printf("│  Author: %s\n", mod.Author)
	fmt.Printf("│  %s\n", mod.Description)
	
	if len(mod.Tags) > 0 {
		fmt.Printf("│  Tags: %s\n", strings.Join(mod.Tags, ", "))
	}
	
	if len(mod.Requires) > 0 {
		fmt.Printf("│  Requires: %s\n", strings.Join(mod.Requires, ", "))
	}
	
	fmt.Printf("│  Repository: %s\n", mod.Repository)
	fmt.Printf("└─\n\n")
}

// showModDetails shows detailed information about a specific mod
func showModDetails(registry *Registry, name string) error {
	var found *RegistryMod
	lowerName := strings.ToLower(name)

	for i := range registry.Mods {
		if strings.ToLower(registry.Mods[i].Name) == lowerName {
			found = &registry.Mods[i]
			break
		}
	}

	if found == nil {
		fmt.Printf("Mod not found: %s\n\n", name)
		fmt.Println("Search for mods with: thorium search <keyword>")
		return nil
	}

	// Print detailed information
	fmt.Printf("╔═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("║ %s\n", found.DisplayName)
	fmt.Printf("╠═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("║\n")
	fmt.Printf("║ Name:        %s\n", found.Name)
	fmt.Printf("║ Version:     %s\n", found.Version)
	fmt.Printf("║ Author:      %s\n", found.Author)
	
	fmt.Printf("║\n")
	fmt.Printf("║ Description:\n")
	
	// Word wrap description
	words := strings.Fields(found.Description)
	line := "║   "
	for _, word := range words {
		if len(line)+len(word)+1 > 65 {
			fmt.Println(line)
			line = "║   " + word
		} else {
			if line == "║   " {
				line += word
			} else {
				line += " " + word
			}
		}
	}
	if line != "║   " {
		fmt.Println(line)
	}
	
	fmt.Printf("║\n")
	fmt.Printf("║ Repository:  %s\n", found.Repository)
	
	fmt.Printf("║\n")
	
	if len(found.Tags) > 0 {
		fmt.Printf("║ Tags:        %s\n", strings.Join(found.Tags, ", "))
	}
	
	if len(found.Requires) > 0 {
		fmt.Printf("║ Requires:    %s\n", strings.Join(found.Requires, ", "))
		fmt.Printf("║              (install dependencies first)\n")
	}
	
	fmt.Printf("║\n")
	fmt.Printf("║ Added:       %s\n", found.AddedDate.Format("2006-01-02"))
	fmt.Printf("║ Updated:     %s\n", found.UpdatedDate.Format("2006-01-02"))
	
	fmt.Printf("║\n")
	fmt.Printf("╚═══════════════════════════════════════════════════════════════\n\n")
	
	fmt.Printf("To install this mod:\n")
	fmt.Printf("  thorium get %s\n\n", found.Repository)
	
	if len(found.Requires) > 0 {
		fmt.Printf("Note: This mod requires the following dependencies:\n")
		for _, dep := range found.Requires {
			fmt.Printf("  - %s (install with: thorium search --name %s)\n", dep, dep)
		}
		fmt.Println()
	}

	return nil
}

// listAllTags lists all unique tags used in the registry
func listAllTags(registry *Registry) error {
	tagCount := make(map[string]int)
	
	for _, mod := range registry.Mods {
		for _, tag := range mod.Tags {
			tagCount[strings.ToLower(tag)]++
		}
	}
	
	if len(tagCount) == 0 {
		fmt.Println("No tags found in registry.")
		return nil
	}
	
	// Sort tags alphabetically
	tags := make([]string, 0, len(tagCount))
	for tag := range tagCount {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	
	fmt.Printf("Available tags (%d total):\n\n", len(tags))
	
	// Group tags by category (simple heuristic based on common patterns)
	categories := map[string][]string{
		"Features":   {},
		"Content":    {},
		"Technical":  {},
		"Gameplay":   {},
		"Scope":      {},
		"Uncategorized": {},
	}
	
	featureTags := []string{"scripting", "database", "ui", "graphics", "audio", "networking", "api"}
	contentTags := []string{"quests", "items", "spells", "npcs", "zones", "dungeons", "raids", "pvp", "professions"}
	technicalTags := []string{"lua-api", "client-server", "client-only", "server-only", "binary-patch", "core-patch"}
	gameplayTags := []string{"balance", "quality-of-life", "convenience", "hardcore", "custom-class", "custom-race"}
	scopeTags := []string{"framework", "library", "content-pack", "patch"}
	
	for _, tag := range tags {
		categorized := false
		
		for _, ft := range featureTags {
			if strings.Contains(tag, ft) {
				categories["Features"] = append(categories["Features"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
				categorized = true
				break
			}
		}
		
		if !categorized {
			for _, ct := range contentTags {
				if strings.Contains(tag, ct) {
					categories["Content"] = append(categories["Content"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
					categorized = true
					break
				}
			}
		}
		
		if !categorized {
			for _, tt := range technicalTags {
				if strings.Contains(tag, tt) {
					categories["Technical"] = append(categories["Technical"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
					categorized = true
					break
				}
			}
		}
		
		if !categorized {
			for _, gt := range gameplayTags {
				if strings.Contains(tag, gt) {
					categories["Gameplay"] = append(categories["Gameplay"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
					categorized = true
					break
				}
			}
		}
		
		if !categorized {
			for _, st := range scopeTags {
				if strings.Contains(tag, st) {
					categories["Scope"] = append(categories["Scope"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
					categorized = true
					break
				}
			}
		}
		
		if !categorized {
			categories["Uncategorized"] = append(categories["Uncategorized"], fmt.Sprintf("%s (%d)", tag, tagCount[tag]))
		}
	}
	
	// Print categories
	categoryOrder := []string{"Features", "Content", "Technical", "Gameplay", "Scope", "Uncategorized"}
	for _, category := range categoryOrder {
		if len(categories[category]) > 0 {
			fmt.Printf("%s:\n", category)
			for _, tag := range categories[category] {
				fmt.Printf("  - %s\n", tag)
			}
			fmt.Println()
		}
	}
	
	fmt.Printf("Search by tag: thorium search --tag <tag-name>\n")
	fmt.Printf("Example:       thorium search --tag networking\n")
	fmt.Printf("Multiple tags: thorium search --tag networking --tag framework\n")
	
	return nil
}

// init registers the command (not used in current architecture, but useful for documentation)
func init() {
	// Command is registered in main.go
}
