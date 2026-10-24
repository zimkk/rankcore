package setup

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func withHome(t *testing.T) string {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestDetectAgentsUsesConfigDir(t *testing.T) {
	home := withHome(t)
	reg := &Registry{Agents: []AgentConfig{
		{ID: "claude-code", Paths: []string{"~/.claude/skills/rank"}},
		{ID: "cursor", Paths: []string{"~/.cursor/skills/rank"}},
	}}
	// A fresh agent install has ~/.claude but no skills/ directory yet.
	os.MkdirAll(filepath.Join(home, ".claude"), 0755)

	got := DetectAgents(reg)
	if len(got) != 1 || got[0].ID != "claude-code" {
		t.Fatalf("expected claude-code detected from ~/.claude, got %+v", got)
	}
}

func TestInstallSkillIsIdempotentAndSafe(t *testing.T) {
	home := withHome(t)
	assets := fstest.MapFS{
		"skill/rank/SKILL.md":         {Data: []byte("---\nname: rank\n---\n# RankCore\n")},
		"skill/rank/workflows/fix.md": {Data: []byte("fix")},
	}
	agent := AgentConfig{ID: "x", Paths: []string{"~/.x/skills/rank"}}
	dir := filepath.Join(home, ".x", "skills", "rank")

	if err := InstallSkill(agent, assets, false); err != nil {
		t.Fatal(err)
	}
	stale := filepath.Join(dir, "workflows", "old.md")
	os.WriteFile(stale, []byte("stale"), 0644)

	if err := InstallSkill(agent, assets, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("reinstall should drop files no longer shipped")
	}
	if !IsRankSkillDir(dir) {
		t.Errorf("installed dir should be recognized as a RankCore skill")
	}

	// Never overwrite a foreign directory.
	foreign := AgentConfig{ID: "y", Paths: []string{"~/.y/skills/rank"}}
	fdir := filepath.Join(home, ".y", "skills", "rank")
	os.MkdirAll(fdir, 0755)
	os.WriteFile(filepath.Join(fdir, "SKILL.md"), []byte("someone else's skill"), 0644)
	if err := InstallSkill(foreign, assets, false); err == nil {
		t.Errorf("expected refusal to overwrite a non-RankCore skill")
	}
}

func TestExpandHome(t *testing.T) {
	home := withHome(t)
	if got := ExpandHome("~/.claude/skills/rank"); got != filepath.Join(home, ".claude", "skills", "rank") {
		t.Errorf("ExpandHome = %s", got)
	}
	if got := ExpandHome("/opt/~weird"); got != filepath.FromSlash("/opt/~weird") {
		t.Errorf("only a leading ~ should expand, got %s", got)
	}
}
