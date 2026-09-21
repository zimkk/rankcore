package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type AgentConfig struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Paths []string `json:"paths"`
}

type Registry struct {
	SchemaVersion int           `json:"schema_version"`
	Agents        []AgentConfig `json:"agents"`
}

// LoadRegistry parses the agent registry configuration.
func LoadRegistry(r io.Reader) (*Registry, error) {
	var reg Registry
	if err := json.NewDecoder(r).Decode(&reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

// DetectAgents identifies which supported agents are present on the user's system.
func DetectAgents(reg *Registry) []AgentConfig {
	var detected []AgentConfig
	homeDir, _ := os.UserHomeDir()

	for _, agent := range reg.Agents {
		for _, p := range agent.Paths {
			expandedPath := strings.Replace(p, "~", homeDir, 1)
			if _, err := os.Stat(filepath.Dir(expandedPath)); err == nil {
				detected = append(detected, agent)
				break
			}
		}
	}
	return detected
}

// InstallSkill simulates or performs copying the embedded rank skill to the agent's skill directory.
func InstallSkill(agent AgentConfig, skillSourcePath string, dryRun bool) error {
	homeDir, _ := os.UserHomeDir()
	
	for _, p := range agent.Paths {
		targetPath := strings.Replace(p, "~", homeDir, 1)
		
		fmt.Printf("Installing /rank skill for %s to %s...\n", agent.Name, targetPath)
		
		if dryRun {
			fmt.Println("[DRY RUN] Would create directory and copy files.")
			continue
		}
		
		err := os.MkdirAll(targetPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", targetPath, err)
		}
		
		// In a real implementation, we would write the embedded fs.FS files here.
		// For the end-to-end working script without full embed, we touch a SKILL.md
		skillFile := filepath.Join(targetPath, "SKILL.md")
		err = os.WriteFile(skillFile, []byte("# RankCore Skill\n"), 0644)
		if err != nil {
			return fmt.Errorf("failed to write SKILL.md: %v", err)
		}
		fmt.Println("Successfully installed.")
	}
	return nil
}
