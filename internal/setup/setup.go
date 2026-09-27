package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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

// CopyFSToDir extracts files from an embedded fs.FS subpath to a target directory on disk.
func CopyFSToDir(src fs.FS, srcSub string, destDir string) error {
	sub, err := fs.Sub(src, srcSub)
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		destPath := filepath.Join(destDir, filepath.FromSlash(path))
		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}
		data, err := fs.ReadFile(sub, path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0644)
	})
}

// InstallSkill copies the embedded rank skill to the agent's skill directory.
func InstallSkill(agent AgentConfig, assetsFS fs.FS, dryRun bool) error {
	homeDir, _ := os.UserHomeDir()

	for _, p := range agent.Paths {
		targetPath := strings.Replace(p, "~", homeDir, 1)

		if dryRun {
			fmt.Printf("  [DRY RUN] Would install /rank to %s\n", targetPath)
			continue
		}

		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
		}

		if assetsFS != nil {
			if err := CopyFSToDir(assetsFS, "skill/rank", targetPath); err != nil {
				return fmt.Errorf("failed to copy skill files to %s: %w", targetPath, err)
			}
		} else {
			skillFile := filepath.Join(targetPath, "SKILL.md")
			if err := os.WriteFile(skillFile, []byte("# RankCore Skill\n"), 0644); err != nil {
				return fmt.Errorf("failed to write SKILL.md: %w", err)
			}
		}
	}
	return nil
}
