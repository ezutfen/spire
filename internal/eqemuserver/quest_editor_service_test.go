package eqemuserver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EQEmu/spire/internal/logger"
	"github.com/EQEmu/spire/internal/pathmgmt"
)

func newQuestEditorTestService(t *testing.T) (*QuestEditorService, string) {
	t.Helper()

	serverRoot := t.TempDir()
	questsDir := filepath.Join(serverRoot, "quests")
	if err := os.MkdirAll(questsDir, 0755); err != nil {
		t.Fatalf("mkdir quests: %v", err)
	}

	t.Setenv("APP_ENV", "local")

	pm := pathmgmt.NewPathManagement(logger.NewAppLogger())
	pm.SetServerPath(serverRoot)
	return NewQuestEditorService(logger.NewAppLogger(), pm, nil), questsDir
}

func TestQuestEditorGetTreeLuaOnly(t *testing.T) {
	service, questsDir := newQuestEditorTestService(t)

	if err := os.MkdirAll(filepath.Join(questsDir, "qeynos"), 0755); err != nil {
		t.Fatalf("mkdir zone: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(questsDir, ".git"), 0755); err != nil {
		t.Fatalf("mkdir hidden: %v", err)
	}
	if err := os.WriteFile(filepath.Join(questsDir, "qeynos", "npc.lua"), []byte("print('ok')"), 0644); err != nil {
		t.Fatalf("write lua: %v", err)
	}
	if err := os.WriteFile(filepath.Join(questsDir, "qeynos", "npc.pl"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("write pl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(questsDir, ".git", "config.lua"), []byte("ignored"), 0644); err != nil {
		t.Fatalf("write hidden lua: %v", err)
	}

	tree, err := service.GetTree()
	if err != nil {
		t.Fatalf("GetTree() error = %v", err)
	}

	if len(tree) != 2 {
		t.Fatalf("expected zone directory and lua file only, got %d entries: %#v", len(tree), tree)
	}
	if !tree[0].IsDirectory || tree[0].Path != "qeynos" {
		t.Fatalf("expected first entry to be qeynos directory, got %#v", tree[0])
	}
	if tree[1].Path != filepath.Join("qeynos", "npc.lua") {
		t.Fatalf("expected lua file entry, got %#v", tree[1])
	}
}

func TestQuestEditorRejectsUnsafePaths(t *testing.T) {
	service, questsDir := newQuestEditorTestService(t)

	if err := os.WriteFile(filepath.Join(questsDir, "safe.lua"), []byte("print('safe')"), 0644); err != nil {
		t.Fatalf("write safe file: %v", err)
	}
	outsideFile := filepath.Join(t.TempDir(), "outside.lua")
	if err := os.WriteFile(outsideFile, []byte("print('outside')"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outsideFile, filepath.Join(questsDir, "escape.lua")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	cases := []string{
		"../outside.lua",
		".hidden.lua",
		"escape.lua",
		"safe.pl",
	}

	for _, path := range cases {
		if _, err := service.GetFile(path); err == nil {
			t.Fatalf("expected GetFile(%q) to fail", path)
		}
	}
}

func TestQuestEditorMoveIntoDirectoryPreservesLuaName(t *testing.T) {
	service, questsDir := newQuestEditorTestService(t)

	if err := os.WriteFile(filepath.Join(questsDir, "source.lua"), []byte("print('x')"), 0644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(questsDir, "dest"), 0755); err != nil {
		t.Fatalf("mkdir dest: %v", err)
	}

	if err := service.MovePath(MovePathRequest{
		OldPath: "source.lua",
		NewPath: "dest",
	}); err != nil {
		t.Fatalf("MovePath() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(questsDir, "dest", "source.lua")); err != nil {
		t.Fatalf("expected moved file in destination directory: %v", err)
	}
}
