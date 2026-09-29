package main

import (
	"os"
	"path/filepath"
	"testing"

	"skillshare/internal/projectdir"
)

func TestResolveAutoMode(t *testing.T) {
	withProject := t.TempDir()
	cfgDir := filepath.Join(withProject, projectdir.Default)
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, projectdir.ConfigFileName), []byte("targets: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	withoutProject := t.TempDir()

	tests := []struct {
		name string
		mode runMode
		cwd  string
		want runMode
	}{
		{"auto with project config", modeAuto, withProject, modeProject},
		{"auto without project config", modeAuto, withoutProject, modeGlobal},
		{"explicit project ignores missing config", modeProject, withoutProject, modeProject},
		{"explicit global ignores project config", modeGlobal, withProject, modeGlobal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveAutoMode(tt.mode, tt.cwd); got != tt.want {
				t.Errorf("resolveAutoMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
