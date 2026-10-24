package setup

import (
	"bytes"
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
	// Detect lists paths whose existence means the agent is installed. When
	// empty, the agent's home config directory (e.g. ~/.claude) is used.
	Detect []string `json:"detect,omitempty"`
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

// ExpandHome replaces a leading "~" with the user's home directory.
func ExpandHome(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return filepath.FromSlash(p)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, filepath.FromSlash(strings.TrimPrefix(p, "~")))
}

// homeConfigDir returns the first directory under ~ for a path like
// "~/.claude/skills/rank" (i.e. "~/.claude").
func homeConfigDir(p string) string {
	rest := strings.TrimPrefix(p, "~/")
	if rest == p {
		return ""
	}
	first := strings.SplitN(rest, "/", 2)[0]
	return "~/" + first
}

// DetectAgents identifies which supported agents are present on the user's system.
func DetectAgents(reg *Registry) []AgentConfig {
	var detected []AgentConfig
	for _, agent := range reg.Agents {
		probes := agent.Detect
		if len(probes) == 0 {
			for _, p := range agent.Paths {
				if d := homeConfigDir(p); d != "" {
					probes = append(probes, d)
				}
			}
		}
		for _, p := range probes {
			if _, err := os.Stat(ExpandHome(p)); err == nil {
				detected = append(detected, agent)
				break
			}
		}
	}
	return detected
}

// Writable reports whether dir (or its nearest existing ancestor) accepts new files.
func Writable(dir string) bool {
	for d := dir; ; d = filepath.Dir(d) {
		if fi, err := os.Stat(d); err == nil {
			if !fi.IsDir() {
				return false
			}
			f, err := os.CreateTemp(d, ".rankcore-probe-*")
			if err != nil {
				return false
			}
			name := f.Name()
			f.Close()
			os.Remove(name)
			return true
		}
		if parent := filepath.Dir(d); parent == d {
			return false
		}
	}
}

// IsRankSkillDir reports whether dir contains a RankCore rank skill, so
// uninstall never deletes a directory it did not create.
func IsRankSkillDir(dir string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return false
	}
	return bytes.Contains(data, []byte("name: rank")) && bytes.Contains(data, []byte("RankCore"))
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
// Re-running it replaces the previous copy, so files dropped from a newer
// release do not linger.
func InstallSkill(agent AgentConfig, assetsFS fs.FS, dryRun bool) error {
	for _, p := range agent.Paths {
		targetPath := ExpandHome(p)

		if dryRun {
			fmt.Printf("  [DRY RUN] Would install /rank to %s\n", targetPath)
			continue
		}

		if _, err := os.Stat(targetPath); err == nil {
			if !IsRankSkillDir(targetPath) {
				return fmt.Errorf("%s exists and is not a RankCore skill; refusing to overwrite", targetPath)
			}
			if err := os.RemoveAll(targetPath); err != nil {
				return fmt.Errorf("failed to replace %s: %w", targetPath, err)
			}
		}

		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
		}

		if err := CopyFSToDir(assetsFS, "skill/rank", targetPath); err != nil {
			return fmt.Errorf("failed to copy skill files to %s: %w", targetPath, err)
		}
	}
	return nil
}
